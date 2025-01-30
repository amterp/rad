package main

import (
	"fmt"
	"io"
	"os"
	"rls/log"
	"rls/server"
)

var StdErr io.Writer = os.Stderr

func main() {
	fmt.Fprintln(StdErr, "Spinning up RSL LSP server...")

	fmt.Fprintln(StdErr, "Initializing logger...")
	log.InitLogger(StdErr)
	log.L.Info("Logger initialized")

	log.L.Info("Creating server...")
	s := server.NewServer(os.Stdin, os.Stdout)

	log.L.Info("Running server...")
	err := s.Run()
	if err != nil {
		log.L.Fatalf("Error running server: %v", err)
	}
	log.L.Info("Exiting...")
}
