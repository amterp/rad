package com

import (
	"fmt"
	"os"
	"strings"

	"github.com/mattn/go-isatty"
)

var (
	IsTty          = checkTty()
	TerminalIsUtf8 = checkTerminalUtf8()
)

func checkTty() bool {
	return isatty.IsTerminal(os.Stdout.Fd())
}

// IsTerminal reports whether f is attached to a terminal.
func IsTerminal(f *os.File) bool {
	return f != nil && isatty.IsTerminal(f.Fd())
}

// ControllingTerminal is the process's controlling terminal, opened directly
// rather than inherited through the standard streams.
//
// In and Out are the same file on Unix, where /dev/tty is one read/write
// device, and distinct on Windows, whose console splits input and output into
// separate devices. Callers must therefore read from In and write to Out rather
// than treating either as the whole terminal.
type ControllingTerminal struct {
	In  *os.File
	Out *os.File

	close func()
}

func (t *ControllingTerminal) Close() {
	t.close()
}

// OpenControllingTerminal opens the process's controlling terminal for reading
// and writing. This is what lets an interactive prompt still reach the user
// when stdin is a pipe or a here-string - notably RED-6's Bash embedding, where
// the script source itself arrives on stdin.
//
// It fails when there is no controlling terminal, which is exactly the case rad
// wants to detect: CI, cron, and agent-driven runs.
func OpenControllingTerminal() (*ControllingTerminal, error) {
	t, err := openControllingTerminal()
	if err != nil {
		return nil, err
	}
	// Opening can succeed for something that isn't usable for raw-mode input;
	// verify rather than handing the prompt engine a file it can't drive.
	if !IsTerminal(t.In) {
		t.Close()
		return nil, fmt.Errorf("%s is not a terminal", t.In.Name())
	}
	return t, nil
}

func checkTerminalUtf8() bool {
	lang := os.Getenv("LANG")
	ctype := os.Getenv("LC_CTYPE")
	// Check for UTF-8 in LANG or LC_CTYPE environment variables
	return strings.Contains(lang, "UTF-8") || strings.Contains(ctype, "UTF-8")
}
