package testing

import (
	"strings"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Ctrl+C during evaluation is the one REPL behavior a snapshot cannot express.
// By the time a statement is running the editor has already given the terminal
// back, so the interrupt arrives as a real signal rather than a keystroke -
// which is exactly why it can reach a running statement at all.

// The session must survive an interrupt. Before the abort policy this exited
// the process, taking every variable with it.
func TestReplSignals_InterruptReturnsToThePrompt(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	stdOut, stdErr := runReplStdin(t, "x = 7\nsleep(30)\nx\nprint(\"still here\")\n")

	assert.Contains(t, stdErr.String(), "Interrupted.",
		"the interrupted statement should say so rather than vanishing")
	assert.Contains(t, stdOut.String(), "7",
		"the session's variables should outlive the interrupt")
	assert.Contains(t, stdOut.String(), "still here",
		"later turns should still evaluate")
	assert.Nil(t, errorOrExit.exitCode, "the process should not have exited")
}

// A turn interrupted part-way still runs the deferred blocks it registered,
// the same way an interrupted script does.
func TestReplSignals_InterruptStillRunsDefers(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	script := "fn work():\n    defer print(\"cleanup\")\n    sleep(30)\n\nwork()\nprint(\"after\")\n"
	stdOut, _ := runReplStdin(t, script)

	out := stdOut.String()
	assert.Contains(t, out, "cleanup")
	assert.Contains(t, out, "after")
	assert.Less(t, strings.Index(out, "cleanup"), strings.Index(out, "after"),
		"the defer belongs to the interrupted turn, so it runs before the next one")
	assert.Nil(t, errorOrExit.exitCode)
}

// runReplStdin drives a REPL session from piped input, which is the path a
// session takes with no terminal to draw on.
func runReplStdin(t *testing.T, stdin string) (stdOut, stdErr *strings.Builder) {
	t.Helper()
	setupAndRun(t, NewTestParams("", "repl", "--color=never").
		StdinInput(stdin).
		NoTerminal())

	var out, err strings.Builder
	out.WriteString(stdOutBuffer.String())
	err.WriteString(stdErrBuffer.String())
	return &out, &err
}
