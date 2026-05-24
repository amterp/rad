package server

import (
	"encoding/json"
	"io"

	"github.com/amterp/rad/radls/analysis"
	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"
)

type Server struct {
	m *Mux
	s *analysis.State
}

func (s *Server) Run() (err error) {
	if err = s.m.Init(); err != nil {
		return
	}

	log.L.Infof("Running mux...")
	err = s.m.Run()
	return
}

func NewServer(r io.Reader, w io.Writer) *Server {
	m := NewMux(r, w)
	server := Server{
		m: m,
		s: analysis.NewState(),
	}

	m.AddRequestHandler(lsp.INITIALIZE, server.handleInitialize)
	m.AddNotificationHandler(lsp.TD_DID_OPEN, server.handleDidOpen)
	m.AddNotificationHandler(lsp.TD_DID_CHANGE, server.handleDidChange)
	m.AddRequestHandler(lsp.TD_COMPLETION, server.handleCompletion)
	m.AddRequestHandler(lsp.TD_CODE_ACTION, server.handleCodeAction)

	return &server
}

func (s *Server) handleInitialize(params json.RawMessage) (result any, err error) {
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

	result = lsp.NewInitializeResult(string(enc))
	return
}

func (s *Server) handleDidOpen(params json.RawMessage) (err error) {
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
		s.notifyDiagnostics(uri, snap.Diagnostics())
	}
	return
}

func (s *Server) handleDidChange(params json.RawMessage) (err error) {
	var didChangeParams lsp.DidChangeTextDocumentParams
	if err = json.Unmarshal(params, &didChangeParams); err != nil {
		return
	}
	uri := didChangeParams.TextDocument.Uri
	s.s.UpdateDoc(uri, didChangeParams.ContentChanges)
	if snap := s.s.Snapshot(uri); snap != nil {
		s.notifyDiagnostics(uri, snap.Diagnostics())
	}
	return
}

func (s *Server) handleCompletion(params json.RawMessage) (result any, err error) {
	var completionParams lsp.CompletionParams
	if err = json.Unmarshal(params, &completionParams); err != nil {
		return
	}
	// Snapshot once at the request boundary, then pass it through.
	// Any subsequent didChange produces a new snapshot but this
	// handler operates on the one it grabbed - frozen, race-free.
	snap := s.s.Snapshot(completionParams.TextDocument.Uri)
	result, err = s.s.Complete(snap, completionParams.Position)
	return
}

func (s *Server) handleCodeAction(params json.RawMessage) (result any, err error) {
	var codeActionParams lsp.CodeActionParams
	if err = json.Unmarshal(params, &codeActionParams); err != nil {
		return
	}
	snap := s.s.Snapshot(codeActionParams.TextDocument.Uri)
	result, err = s.s.CodeAction(snap, codeActionParams.Range)
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
