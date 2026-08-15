package core

import "fmt"

// replDescription describes the `rad repl` command in usage/completion output.
const replDescription = "Starts an interactive REPL session."

// RunRepl is the entry point for REPL mode.
func RunRepl() error {
	session, err := NewReplSession()
	if err != nil {
		return fmt.Errorf("failed to start REPL: %w", err)
	}
	defer func() {
		if err := session.Shutdown(); err != nil {
			RP.RadDebugf("REPL shutdown error: %v", err)
		}
	}()

	return session.Run()
}

func printReplBanner() {
	RP.Printf("🤙 Rad REPL %s\n", Version)
	RP.Printf("Type ':help' for help, ':exit' or Ctrl+D to quit.\n\n")
}
