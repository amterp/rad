//go:build unix

package testing

import (
	"syscall"
	"testing"

	"github.com/amterp/rad/core"
)

// Verifies signal_ignore plumbs through to the OS signal seam for each named
// signal (scalar and list forms). SIGPIPE is the motivating case: a script
// piping its output to e.g. `head` should not die when the consumer closes the
// pipe. We assert the SIG_IGN request reached the seam rather than the
// process-level outcome, which can't be exercised deterministically in-process.
func TestSignals_IgnorePlumbsToOsSeam(t *testing.T) {
	fake := core.NewFakeSignalSource()
	runnerInput.RSignal = fake
	t.Cleanup(func() { runnerInput.RSignal = nil })

	script := `
signal_ignore("sigpipe")
signal_ignore(["sighup", "sigusr2"])
print("done")
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer, "done\n")
	assertExitCode(t, 0)

	for _, sig := range []syscall.Signal{syscall.SIGPIPE, syscall.SIGHUP, syscall.SIGUSR2} {
		if !fake.WasIgnored(sig) {
			t.Errorf("Expected signal_ignore to install SIG_IGN for %v, but it didn't", sig)
		}
	}
}
