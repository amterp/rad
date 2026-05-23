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

	assertOutput(t, stdOutBuffer, "starting\ndefer ran\n")
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
	assertOutput(t, stdOutBuffer, "errdefer\nnormal defer\n")
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

	assertOutput(t, stdOutBuffer, "defer ran\n")
	assertExitCode(t, 143)
}

// Verifies signal_trap handler is invoked with a ctx map containing the
// expected signal name and conventional exit code. After the handler returns,
// the script continues (always-continue semantics), and the script's natural
// termination produces exit code 0.
func TestSignals_TrapHandlerRunsWithCtx(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	script := `
signal_trap("sigint", fn(ctx):
    print("sig={ctx.signal} code={ctx.exit_code}")
)

sleep(30)
print("after sleep")
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer, "sig=sigint code=130\nafter sleep\n")
	assertExitCode(t, 0)
}

// Verifies that calling exit() inside the handler produces the expected exit
// code and triggers defers.
func TestSignals_TrapHandlerExits(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	script := `
defer:
    print("defer ran")

signal_trap("sigint", fn(ctx):
    print("handler")
    exit(ctx.exit_code)
)

sleep(30)
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer, "handler\ndefer ran\n")
	assertExitCode(t, 130)
}

// Verifies that re-registering replaces the previous handler.
func TestSignals_TrapReregisterReplaces(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	script := `
signal_trap("sigint", fn(ctx):
    print("first handler")
)
signal_trap("sigint", fn(ctx):
    print("second handler")
)

sleep(30)
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer, "second handler\n")
	assertExitCode(t, 0)
}

// Verifies that handler args are merely positional - if the handler body
// doesn't reference ctx, the script still runs without error.
func TestSignals_TrapHandlerIgnoresCtx(t *testing.T) {
	setupSignalTest(t, syscall.SIGINT)

	script := `
signal_trap("sigint", fn(ctx):
    print("any handler will do")
)

sleep(30)
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer, "any handler will do\n")
	assertExitCode(t, 0)
}

// Verifies that two distinct signals arriving in quick succession both get
// dispatched, in the order they were delivered. Earlier the manager held a
// single atomic.Pointer for pendingSig which silently overwrote the first
// signal when a second arrived between checkpoints; this test pins down the
// queue behavior.
func TestSignals_QueuesMultipleDistinctSignals(t *testing.T) {
	fake := core.NewFakeSignalSource()
	runnerInput.RSignal = fake

	// Fire both signals from the single sleep call so they both land in the
	// notifyCh buffer before the checkpoint drains either.
	signalingSleep := func(ctx context.Context, dur time.Duration) {
		fake.Fire(syscall.SIGTERM)
		fake.Fire(syscall.SIGINT)
		time.Sleep(20 * time.Millisecond)
	}
	runnerInput.RSleep = &signalingSleep
	t.Cleanup(func() {
		runnerInput.RSignal = nil
		baseline := newRunnerInput()
		runnerInput.RSleep = baseline.RSleep
	})

	script := `
signal_trap(["sigint", "sigterm"], fn(ctx):
    print("got {ctx.signal}")
)

sleep(30)
`
	setupAndRunCode(t, script, "--color=never")

	// Both handlers must run. Order reflects delivery order (SIGTERM first,
	// then SIGINT) - the queue is FIFO.
	assertOutput(t, stdOutBuffer, "got sigterm\ngot sigint\n")
	assertExitCode(t, 0)
}

// Verifies a multi-signal handler dispatches with the correct name on the
// ctx for each signal type it serves.
func TestSignals_TrapMultiSignalReceivesCorrectName(t *testing.T) {
	setupSignalTest(t, syscall.SIGTERM)

	script := `
signal_trap(["sigint", "sigterm"], fn(ctx):
    print("got {ctx.signal}")
)

sleep(30)
`
	setupAndRunCode(t, script, "--color=never")

	assertOutput(t, stdOutBuffer, "got sigterm\n")
	assertExitCode(t, 0)
}
