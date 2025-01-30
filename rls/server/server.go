package server

import (
	"encoding/json"
	"io"
	"rls/analysis"
	"rls/log"
	"rls/lsp"
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
	log.L.Infof("Received initialize from %s %s", initParams.ClientInfo.Name, initParams.ClientInfo.Version)
	result = lsp.NewInitializeResult()
	return
}

func (s *Server) handleDidOpen(params json.RawMessage) (err error) {
	var didOpenParams lsp.DidOpenTextDocumentParams
	if err = json.Unmarshal(params, &didOpenParams); err != nil {
		return
	}
	s.s.AddDoc(didOpenParams.TextDocument.Uri, didOpenParams.TextDocument.Text)
	return
}

func (s *Server) handleDidChange(params json.RawMessage) (err error) {
	var didChangeParams lsp.DidChangeTextDocumentParams
	if err = json.Unmarshal(params, &didChangeParams); err != nil {
		return
	}
	s.s.UpdateDoc(didChangeParams.TextDocument.Uri, didChangeParams.ContentChanges)
	return
}

func (s *Server) handleCompletion(params json.RawMessage) (result any, err error) {
	var completionParams lsp.CompletionParams
	if err = json.Unmarshal(params, &completionParams); err != nil {
		return
	}
	result, err = s.s.Complete(completionParams.TextDocument.Uri, completionParams.Position)
	return
}

func (s *Server) handleCodeAction(params json.RawMessage) (result any, err error) {
	var codeActionParams lsp.CodeActionParams
	if err = json.Unmarshal(params, &codeActionParams); err != nil {
		return
	}
	result, err = s.s.CodeAction(codeActionParams.TextDocument.Uri, codeActionParams.Range)
	return
}
