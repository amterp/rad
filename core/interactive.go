package core

import (
	"os"

	"github.com/amterp/radish"
)

// InteractiveDriver runs a radish prompt Model to completion. Production wraps the
// real terminal; tests inject a scripted driver (radish.ScriptDriver) so the real
// prompt logic and rendering run end-to-end without a TTY.
//
// The no-TTY policy lives here, not in radish: the production driver delegates to
// radish.RunTerminal, which returns radish.ErrNotInteractive when stdin is not a
// terminal. Callers map that to a clear, actionable error. A scripted driver never
// reports ErrNotInteractive, which is exactly why interactive prompts become
// testable.
type InteractiveDriver interface {
	Run(model radish.Model) (radish.Result, radish.Model, error)
}

// terminalDriver is the production driver: it reads keystrokes from stdin in raw
// mode and renders the prompt to stderr, keeping stdout clean for the script's
// own output.
type terminalDriver struct{}

func (terminalDriver) Run(model radish.Model) (radish.Result, radish.Model, error) {
	return radish.RunTerminal(model, os.Stdin, RIo.StdErr)
}
