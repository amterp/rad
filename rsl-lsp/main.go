package main

import (
	"fmt"
	"os"
	"rsl-lsp/log"
	"rsl-lsp/server"
)

func main() {
	fmt.Println("Spinning up RSL LSP server...")

	fmt.Println("Initializing logger...")
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
