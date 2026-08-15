package core

import (
	"fmt"
	"io"
)

// replDescription describes the `rad repl` command in usage/completion output.
const replDescription = "Starts an interactive REPL session."

// CreateReplSession creates a new REPL session with default configuration
func CreateReplSession() (ReplSession, error) {
	// Skeleton implementation
	interpreter := NewInterpreter(InterpreterInput{
		Src:            "",
		Tree:           nil,
		ScriptName:     "repl",
		InvokedCommand: nil,
	})

	interpreter.InitBuiltIns()

	inputReader := NewSingleLineInputReader()
	session := NewReplSession(interpreter, inputReader)

	return session, nil
}

// RunRepl is the main entry point for REPL mode
func RunRepl() error {
	session, err := CreateReplSession()
	if err != nil {
		return fmt.Errorf("failed to create REPL session: %w", err)
	}

	// A REPL hosts execution rather than being it, so a fatal error unwinds to
	// the prompt instead of ending the process.
	RExit.SetUnwinding(true)
	defer RExit.SetUnwinding(false)

	defer func() {
		if shutdownErr := session.Shutdown(); shutdownErr != nil {
			RP.Printf("Warning: REPL shutdown error: %v\n", shutdownErr)
		}
	}()

	return session.Run()
}

// shouldPrintResult determines if an execution result should be auto-printed
func shouldPrintResult(result *ExecutionResult) bool {
	if result.Error != nil {
		return false // Errors are handled separately
	}

	if result.Value == VOID_SENTINEL {
		return false // No value to print
	}

	// Use the ShouldPrint flag determined during execution
	return result.ShouldPrint
}

// printWelcomeBanner prints the REPL welcome message
func printWelcomeBanner() {
	RP.Printf("🤙 Rad REPL %s\n", Version)
	RP.Printf("Type 'exit()' to quit.\n\n")
}

// handleReplError handles and displays REPL-specific errors
func handleReplError(err error) {
	if err == io.EOF {
		// Ctrl+D - clean exit
		RP.Printf("\n")
		return
	}

	// Format error message appropriately
	RP.Printf("Error: %v\n", err)
}
