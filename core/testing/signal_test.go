package testing

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/amterp/rad/core"
)

// setupSignalTest swaps in a FakeSignalSource and a sleep mock that fires a
// signal on entry. Returns the fake so the test can run assertions on it.
// The sleep mock does a brief time.Sleep so the SignalManager goroutine has
// a chance to drain the channel before sleep returns and the next checkpoint
// runs. Twenty ms is well above any reasonable scheduling delay; if tests
// turn flaky, bump it or add explicit sync.
func setupSignalTest(t *testing.T, fireOnSleep syscall.Signal) *core.FakeSignalSource {
	t.Helper()
	fake := core.NewFakeSignalSource()
	runnerInput.RSignal = fake

	signalingSleep := func(ctx context.Context, dur time.Duration) {
		fake.Fire(fireOnSleep)
		time.Sleep(20 * time.Millisecond)
	}
	runnerInput.RSleep = &signalingSleep

	t.Cleanup(func() {
		runnerInput.RSignal = nil
		baseline := newRunnerInput()
		runnerInput.RSleep = baseline.RSleep
	})
	return fake
}

// Verifies that a SIGINT arriving during a sleep causes the deferred block
// to run before exit, and that the exit code reflects the signal (130).
func TestSignals_DeferRunsOnSigint(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	script := `
defer:
    print("defer ran")

print("starting")
sleep(30)
print("after sleep")
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer,"starting\ndefer ran\n")
	assertExitCode(t, 130)
}

// Verifies errdefer also runs (because the signal-triggered exit code is
// non-zero - 130 for SIGINT - so errdefer is not skipped).
func TestSignals_ErrdeferRunsOnSigint(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	script := `
defer:
    print("normal defer")

errdefer:
    print("errdefer")

sleep(30)
`
	setupAndRunCode(t, script, "--color=never")

	// LIFO ordering: errdefer registered second runs first.
	assertOutput(t, stdOutBuffer,"errdefer\nnormal defer\n")
	assertExitCode(t, 130)
}

// Verifies SIGTERM produces exit code 143 (128 + 15) and that defers run.
func TestSignals_DeferRunsOnSigterm(t *testing.T) {
	setupSignalTest(t, syscall.SIGTERM)

	script := `
defer:
    print("defer ran")

sleep(30)
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer,"defer ran\n")
	assertExitCode(t, 143)
}
