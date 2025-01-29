package server

import (
	"encoding/json"
	"io"
	"rsl-lsp/log"
	"rsl-lsp/lsp"
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
	var initializeParams lsp.InitializeParams
	if err = json.Unmarshal(params, &initializeParams); err != nil {
		return
	}
	result = lsp.NewInitializeResult()
	return
}
