package core

import (
	"errors"
	"io"
	"os"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/radish"
)

// InteractiveDriver runs a radish prompt Model to completion. Production wraps the
// real terminal; tests inject a scripted driver (radish.ScriptDriver) so the real
// prompt logic and rendering run end-to-end without a TTY.
//
// The no-TTY policy lives here, not in radish: the production driver resolves a
// terminal through RTerminal and reports radish.ErrNotInteractive when there
// isn't one. Callers map that to a clear, actionable error. A scripted driver
// never reports ErrNotInteractive, which is exactly why interactive prompts
// become testable.
type InteractiveDriver interface {
	Run(model radish.Model) (radish.Result, radish.Model, error)
}

// ErrNoTerminal reports that no terminal is reachable: neither stdin nor the
// process's controlling terminal. This is the CI / cron / AI-agent case.
var ErrNoTerminal = errors.New("no terminal available")

// Terminal is a resolved place to prompt: somewhere to read keystrokes from and
// somewhere to draw. Release it when the prompt finishes.
type Terminal struct {
	In    *os.File
	Out   io.Writer
	close func()
}

func (t *Terminal) Release() {
	if t != nil && t.close != nil {
		t.close()
	}
}

// TerminalSource decides where interactive prompts talk to the user. It is the
// single authority on "can rad prompt at all" - both the production driver and
// the pre-flight guard ask it, so the two can never disagree.
//
// Note com.IsTty is NOT this: it inspects stdout, which answers a different
// question and is only used for cosmetic rendering choices.
type TerminalSource interface {
	Open() (*Terminal, error)
}

// osTerminalSource prefers stdin, then falls back to the controlling terminal.
// The fallback is what lets a prompt reach the user when stdin carries data
// rather than keystrokes - `echo x | rad s.rad`, and RED-6's Bash embedding,
// where the script source itself arrives on stdin.
type osTerminalSource struct{}

func (osTerminalSource) Open() (*Terminal, error) {
	// Drawing on stderr is right only when stderr is itself a terminal. Under
	// `rad s.rad 2>err.log` it would write the prompt into the file and leave
	// the user typing blind, so a redirected stderr falls through to the
	// controlling terminal for output even when stdin is fine for input.
	if com.IsTerminal(os.Stdin) && stdErrIsTerminal() {
		return &Terminal{In: os.Stdin, Out: RIo.StdErr}, nil
	}

	tty, err := com.OpenControllingTerminal()
	if err != nil {
		if com.IsTerminal(os.Stdin) {
			// Keystrokes still reach us; only the drawing surface is redirected.
			// A prompt the user can't see beats no prompt at all.
			return &Terminal{In: os.Stdin, Out: RIo.StdErr}, nil
		}
		return nil, ErrNoTerminal
	}

	in := os.Stdin
	if !com.IsTerminal(in) {
		in = tty.In
	}
	out := RIo.StdErr
	if !stdErrIsTerminal() {
		out = tty.Out
	}

	return &Terminal{In: in, Out: out, close: tty.Close}, nil
}

func stdErrIsTerminal() bool {
	f, ok := RIo.StdErr.(*os.File)
	return ok && com.IsTerminal(f)
}

// fixedTerminalSource reports a canned answer, so tests can decide whether a
// terminal exists instead of inheriting whatever `go test` was launched with.
type fixedTerminalSource struct {
	available bool
}

// NewFakeTerminalSource returns a TerminalSource that reports terminal
// availability without touching a real one.
func NewFakeTerminalSource(available bool) TerminalSource {
	return fixedTerminalSource{available: available}
}

func (f fixedTerminalSource) Open() (*Terminal, error) {
	if !f.available {
		return nil, ErrNoTerminal
	}
	// In is nil: tests that get here have replaced the driver, so nothing reads it.
	return &Terminal{Out: io.Discard}, nil
}

// TerminalAvailable reports whether an interactive prompt could reach a user.
func TerminalAvailable() bool {
	term, err := RTerminal.Open()
	if err != nil {
		return false
	}
	term.Release()
	return true
}

// terminalDriver is the production driver: it reads keystrokes in raw mode from
// the resolved terminal and renders the prompt away from stdout, keeping stdout
// clean for the script's own output.
type terminalDriver struct{}

func (terminalDriver) Run(model radish.Model) (radish.Result, radish.Model, error) {
	term, err := RTerminal.Open()
	if err != nil {
		return radish.Result{}, model, radish.ErrNotInteractive
	}
	defer term.Release()
	return radish.RunTerminal(model, term.In, term.Out)
}
