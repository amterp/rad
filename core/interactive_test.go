package core

import (
	"bytes"
	"os"
	"testing"

	"github.com/amterp/radish"
	"github.com/stretchr/testify/assert"
)

// withTerminal swaps the terminal seam for one test and restores it after.
func withTerminal(t *testing.T, src TerminalSource) {
	t.Helper()
	prev := RTerminal
	RTerminal = src
	t.Cleanup(func() { RTerminal = prev })
}

// The whole no-TTY policy rests on this mapping: every prompt call site checks
// errors.Is(err, radish.ErrNotInteractive) to decide it cannot ask the user.
func TestTerminalDriverReportsNotInteractiveWhenNoTerminal(t *testing.T) {
	withTerminal(t, NewFakeTerminalSource(false))

	_, _, err := terminalDriver{}.Run(radish.NewInput())

	assert.ErrorIs(t, err, radish.ErrNotInteractive)
}

func TestTerminalAvailableFollowsTheSeam(t *testing.T) {
	withTerminal(t, NewFakeTerminalSource(false))
	assert.False(t, TerminalAvailable())

	withTerminal(t, NewFakeTerminalSource(true))
	assert.True(t, TerminalAvailable())
}

func TestFakeTerminalSourceOpenIsReleasable(t *testing.T) {
	term, err := NewFakeTerminalSource(true).Open()
	assert.NoError(t, err)
	assert.NotNil(t, term)
	term.Release() // must not panic despite having no close func

	_, err = NewFakeTerminalSource(false).Open()
	assert.ErrorIs(t, err, ErrNoTerminal)
}

// A redirected stderr must not be mistaken for somewhere the user can see the
// prompt - that is what sends osTerminalSource to draw on /dev/tty instead.
func TestStdErrIsTerminalRejectsNonTerminals(t *testing.T) {
	prev := RIo
	t.Cleanup(func() { RIo = prev })

	RIo = RadIo{StdErr: &bytes.Buffer{}}
	assert.False(t, stdErrIsTerminal(), "a buffer is not a terminal")

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	assert.NoError(t, err)
	t.Cleanup(func() { devNull.Close() })

	RIo = RadIo{StdErr: devNull}
	assert.False(t, stdErrIsTerminal(), "os.DevNull is a file but not a terminal")
}
