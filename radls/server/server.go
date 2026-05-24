package server

import (
	"context"
	"encoding/json"
	"io"
	"time"

	"github.com/amterp/rad/radls/analysis"
	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"
)

// diagnosticsDebounceDelay is the idle window after a didChange before
// we publish diagnostics. Long enough that burst typing doesn't flicker
// errors, short enough that the user sees them when they pause.
const diagnosticsDebounceDelay = 200 * time.Millisecond

type Server struct {
	m       *Mux
	s       *analysis.State
	diagDeb *Debouncer
}

func (s *Server) Run() (err error) {
	if err = s.m.Init(); err != nil {
		return
	}

	log.L.Infof("Running mux...")
	err = s.m.Run()
	return
}

// NewServer creates a server using the default diagnostics debounce
// delay. Tests that need synchronous publish should use
// NewServerWithDebounce(r, w, 0) instead.
func NewServer(r io.Reader, w io.Writer) *Server {
	return NewServerWithDebounce(r, w, diagnosticsDebounceDelay)
}

// NewServerWithDebounce lets tests override the publish-diagnostics
// debounce delay. A zero delay makes Trigger synchronous, which keeps
// the snapshot-test harness deterministic without forcing it to sleep
// for arbitrary windows.
func NewServerWithDebounce(r io.Reader, w io.Writer, delay time.Duration) *Server {
	m := NewMux(r, w)
	server := Server{
		m:       m,
		s:       analysis.NewState(),
		diagDeb: NewDebouncer(delay),
	}

	m.AddRequestHandler(lsp.INITIALIZE, server.handleInitialize)
	m.AddNotificationHandler(lsp.TD_DID_OPEN, server.handleDidOpen)
	m.AddNotificationHandler(lsp.TD_DID_CHANGE, server.handleDidChange)
	m.AddRequestHandler(lsp.TD_COMPLETION, server.handleCompletion)
	m.AddRequestHandler(lsp.TD_CODE_ACTION, server.handleCodeAction)
	m.AddRequestHandler(lsp.TD_HOVER, server.handleHover)
	m.AddRequestHandler(lsp.TD_DEFINITION, server.handleDefinition)
	m.AddRequestHandler(lsp.TD_DOCUMENT_SYMBOL, server.handleDocumentSymbol)
	m.AddRequestHandler(lsp.TD_REFERENCES, server.handleReferences)
	m.AddRequestHandler(lsp.TD_SEMANTIC_TOKENS, server.handleSemanticTokens)

	return &server
}

func (s *Server) handleInitialize(_ context.Context, params json.RawMessage) (result any, err error) {
	var initParams lsp.InitializeParams
	if err = json.Unmarshal(params, &initParams); err != nil {
		return
	}
	clientName, clientVersion := "(unknown)", "(unknown)"
	if initParams.ClientInfo != nil {
		clientName = initParams.ClientInfo.Name
		clientVersion = initParams.ClientInfo.Version
	}
	log.L.Infof("Received initialize from %s %s", clientName, clientVersion)

	var offered []analysis.PositionEncoding
	if initParams.Capabilities.General != nil {
		for _, e := range initParams.Capabilities.General.PositionEncodings {
			offered = append(offered, analysis.PositionEncoding(e))
		}
	}
	enc := analysis.NegotiatePositionEncoding(offered)
	s.s.SetEncoding(enc)
	log.L.Infof("Negotiated position encoding: %s (client offered %v)", enc, offered)

	result = lsp.NewInitializeResult(string(enc), analysis.SemanticTokensLegend())
	return
}

func (s *Server) handleDidOpen(_ context.Context, params json.RawMessage) (err error) {
	var didOpenParams lsp.DidOpenTextDocumentParams
	if err = json.Unmarshal(params, &didOpenParams); err != nil {
		return
	}
	uri := didOpenParams.TextDocument.Uri
	s.s.AddDoc(uri, didOpenParams.TextDocument.Text)
	// Grab the freshly-built snapshot to publish diagnostics from. If
	// somehow the snapshot is gone we skip - notifyDiagnostics with an
	// empty slice would clear any prior diagnostics, which is wrong.
	if snap := s.s.Snapshot(uri); snap != nil {
		defer snap.Release()
		s.notifyDiagnostics(uri, snap.Diagnostics())
	}
	return
}

func (s *Server) handleDidChange(_ context.Context, params json.RawMessage) (err error) {
	var didChangeParams lsp.DidChangeTextDocumentParams
	if err = json.Unmarshal(params, &didChangeParams); err != nil {
		return
	}
	uri := didChangeParams.TextDocument.Uri
	// Analysis runs synchronously here so hover/goto-def sees fresh
	// state. Only the wire publish is debounced - the goal is to
	// suppress per-keystroke flicker, not to delay the analyzer.
	s.s.UpdateDoc(uri, didChangeParams.ContentChanges)
	s.diagDeb.Trigger(uri, func() {
		// Re-grab the snapshot at fire time, not trigger time:
		// further keystrokes between trigger and fire will have
		// produced newer versions, and we want the latest.
		if snap := s.s.Snapshot(uri); snap != nil {
			defer snap.Release()
			s.notifyDiagnostics(uri, snap.Diagnostics())
		}
	})
	return
}

func (s *Server) handleCompletion(_ context.Context, params json.RawMessage) (result any, err error) {
	var completionParams lsp.CompletionParams
	if err = json.Unmarshal(params, &completionParams); err != nil {
		return
	}
	// Snapshot once at the request boundary, then pass it through.
	// Any subsequent didChange produces a new snapshot but this
	// handler operates on the one it grabbed - frozen, race-free.
	snap := s.s.Snapshot(completionParams.TextDocument.Uri)
	if snap != nil {
		defer snap.Release()
	}
	result, err = s.s.Complete(snap, completionParams.Position)
	return
}

func (s *Server) handleCodeAction(_ context.Context, params json.RawMessage) (result any, err error) {
	var codeActionParams lsp.CodeActionParams
	if err = json.Unmarshal(params, &codeActionParams); err != nil {
		return
	}
	snap := s.s.Snapshot(codeActionParams.TextDocument.Uri)
	if snap != nil {
		defer snap.Release()
	}
	result, err = s.s.CodeAction(snap, codeActionParams.Range)
	return
}

func (s *Server) handleHover(_ context.Context, params json.RawMessage) (result any, err error) {
	var hoverParams lsp.HoverParams
	if err = json.Unmarshal(params, &hoverParams); err != nil {
		return
	}
	snap := s.s.Snapshot(hoverParams.TextDocument.Uri)
	if snap != nil {
		defer snap.Release()
	}
	// Returning nil for the result field encodes the LSP-spec "null"
	// reply that clients treat as "no hover here," which is what we
	// want when the cursor isn't on a known identifier. State.Hover
	// returns (nil, nil) for that case so we pass it through.
	result, err = s.s.Hover(snap, hoverParams.Position)
	return
}

func (s *Server) handleDefinition(_ context.Context, params json.RawMessage) (result any, err error) {
	var defParams lsp.DefinitionParams
	if err = json.Unmarshal(params, &defParams); err != nil {
		return
	}
	snap := s.s.Snapshot(defParams.TextDocument.Uri)
	if snap != nil {
		defer snap.Release()
	}
	// Same nil-passthrough story as hover - a nil *lsp.Location
	// marshals as null, which is the spec-defined "no definition
	// found" reply.
	result, err = s.s.Definition(snap, defParams.Position)
	return
}

func (s *Server) handleDocumentSymbol(_ context.Context, params json.RawMessage) (result any, err error) {
	var docSymParams lsp.DocumentSymbolParams
	if err = json.Unmarshal(params, &docSymParams); err != nil {
		return
	}
	snap := s.s.Snapshot(docSymParams.TextDocument.Uri)
	if snap != nil {
		defer snap.Release()
	}
	// Empty slice (not nil) is the right answer here - the LSP wire
	// expects a JSON array; nil would marshal as null and trip some
	// clients that gracefully degrade only for non-arrays.
	result, err = s.s.DocumentSymbols(snap)
	return
}

func (s *Server) handleReferences(_ context.Context, params json.RawMessage) (result any, err error) {
	var refParams lsp.ReferenceParams
	if err = json.Unmarshal(params, &refParams); err != nil {
		return
	}
	snap := s.s.Snapshot(refParams.TextDocument.Uri)
	if snap != nil {
		defer snap.Release()
	}
	result, err = s.s.References(snap, refParams.Position, refParams.Context.IncludeDeclaration)
	return
}

func (s *Server) handleSemanticTokens(_ context.Context, params json.RawMessage) (result any, err error) {
	var stParams lsp.SemanticTokensParams
	if err = json.Unmarshal(params, &stParams); err != nil {
		return
	}
	snap := s.s.Snapshot(stParams.TextDocument.Uri)
	if snap != nil {
		defer snap.Release()
	}
	result, err = s.s.SemanticTokens(snap)
	return
}

func (s *Server) notifyDiagnostics(uri string, diagnostics []lsp.Diagnostic) {
	log.L.Infof("Notifying of %d diagnostics for %s", len(diagnostics), uri)
	err := s.m.Notify(lsp.TD_PUBLISH_DIAGNOSTICS, lsp.NewPublishDiagnosticsParams(uri, diagnostics))
	if err != nil {
		// todo notify client?
		log.L.Errorf("Failed to notify diagnostics: %v", err)
	}
}
