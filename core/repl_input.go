package core

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// SingleLineInputReader implements InputReader for single-line input (MVP)
// Uses RadIo abstractions for full testability
type SingleLineInputReader struct {
	reader             *bufio.Scanner
	primaryPrompt      string
	continuationPrompt string // unused in MVP, but ready for future multi-line support
	writer             io.Writer
}

// NewSingleLineInputReader creates a new single-line input reader using RadIo
func NewSingleLineInputReader() InputReader {
	return &SingleLineInputReader{
		reader:             bufio.NewScanner(RIo.StdIn.Unwrap()),
		primaryPrompt:      "> ",
		continuationPrompt: "... ", // ready for future use
		writer:             RIo.StdOut,
	}
}

// ReadStatement reads a single line of input from the user
func (r *SingleLineInputReader) ReadStatement() (string, error) {
	fmt.Fprint(r.writer, r.primaryPrompt)
	if !r.reader.Scan() {
		// Handle EOF or error
		if err := r.reader.Err(); err != nil {
			return "", err
		}
		// EOF (Ctrl+D) - return special error to signal exit
		return "", io.EOF
	}

	line := strings.TrimSpace(r.reader.Text())
	return line, nil
}

// SupportsMultiLine returns false for MVP (single-line only)
func (r *SingleLineInputReader) SupportsMultiLine() bool {
	return false
}

// SetPrompt allows customizing the prompts
func (r *SingleLineInputReader) SetPrompt(primary, continuation string) {
	r.primaryPrompt = primary
	r.continuationPrompt = continuation
}

// Shutdown cleans up resources
func (r *SingleLineInputReader) Shutdown() error {
	return nil
}

// Future: ReadlineInputReader for enhanced input with history
// This interface design allows us to swap in a readline-based implementation later
type ReadlineInputReader struct {
	// todo: Will implement with github.com/chzyer/readline
	// todo: Command history support
	// todo: Multi-line input support
	// todo: Tab completion hooks
}

// todo: Implement ReadlineInputReader methods when adding enhanced features
