package server

import (
	"encoding/json"
	"io"
	"rls/log"
	"rls/lsp"
)

type Server struct {
	m *Mux
	// todo state
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
	}

	m.AddRequestHandler(lsp.INITIALIZE, server.handleInitialize)

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
