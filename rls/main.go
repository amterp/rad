package main

import (
	"fmt"
	"os"
	"rls/log"
	"rls/server"
)

func main() {
	fmt.Fprintln(os.Stderr, "Spinning up RSL LSP server...")

	fmt.Fprintln(os.Stderr, "Initializing logger...")
	log.InitLogger()
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
