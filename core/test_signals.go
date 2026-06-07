package core

import (
	"os"
	"sync"
)

// FakeSignalSource is a test-only SignalSource that lets tests deliver
// signals deterministically rather than sending real OS signals.
//
// Production code goes through realSignalSource (signals.go), which wraps
// os/signal. Tests inject a FakeSignalSource via RunnerInput.RSignal; calls
// to Notify record the subscription, and Fire delivers the named signal to
// every channel subscribed for it.
//
// Lives in the production package (not _test.go) so test packages outside
// core/ can use it. Not part of the Rad public API.
type FakeSignalSource struct {
	mu      sync.Mutex
	subs    map[os.Signal][]chan<- os.Signal
	fired   []os.Signal // history, useful for assertions
	ignored []os.Signal // SIG_IGN history, useful for assertions
}

func NewFakeSignalSource() *FakeSignalSource {
	return &FakeSignalSource{
		subs: make(map[os.Signal][]chan<- os.Signal),
	}
}

// Notify records that ch wants to receive sigs. Mirrors signal.Notify,
// including its additive-but-idempotent semantics: subscribing the same
// channel+signal twice does NOT cause two deliveries on Fire. Without this
// dedup, Start() subscribes SIGINT then signal_trap("sigint", ...) subscribes
// it again, and a single Fire would deliver SIGINT twice - confusing the
// SignalManager's double-SIGINT detection.
func (f *FakeSignalSource) Notify(ch chan<- os.Signal, sigs ...os.Signal) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, sig := range sigs {
		already := false
		for _, existing := range f.subs[sig] {
			if existing == ch {
				already = true
				break
			}
		}
		if !already {
			f.subs[sig] = append(f.subs[sig], ch)
		}
	}
}

// Stop removes ch from all subscriptions.
func (f *FakeSignalSource) Stop(ch chan<- os.Signal) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for sig, chs := range f.subs {
		kept := chs[:0]
		for _, existing := range chs {
			if existing != ch {
				kept = append(kept, existing)
			}
		}
		f.subs[sig] = kept
	}
}

// Ignore records the SIG_IGN request without touching the real process. Tests
// assert against WasIgnored to verify signal_ignore plumbs through to the OS
// seam (we can't deterministically test "process survives SIGPIPE" in-process).
func (f *FakeSignalSource) Ignore(sigs ...os.Signal) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.ignored = append(f.ignored, sigs...)
}

// WasIgnored reports whether sig was passed to Ignore since the source was
// created. Used by tests asserting signal_ignore behavior.
func (f *FakeSignalSource) WasIgnored(sig os.Signal) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, s := range f.ignored {
		if s == sig {
			return true
		}
	}
	return false
}

// Fire delivers sig to every channel currently subscribed to it.
//
// Writes are non-blocking; if a channel's buffer is full the signal is
// dropped (mirrors how os/signal behaves under burst conditions). The
// SignalManager uses a 4-deep buffer so single-shot Fires are reliable.
//
// Note: signal processing happens asynchronously on the SignalManager
// dispatch goroutine. Tests that need to observe the result should either
// (a) use a control flow that naturally synchronizes (e.g. assert defers ran
// after script exit), or (b) call FireAndWait below.
func (f *FakeSignalSource) Fire(sig os.Signal) {
	f.mu.Lock()
	chs := f.subs[sig]
	f.fired = append(f.fired, sig)
	f.mu.Unlock()

	for _, ch := range chs {
		select {
		case ch <- sig:
		default:
		}
	}
}

// FiredCount returns how many times sig has been fired since the source
// was created. Useful for assertions in tests.
func (f *FakeSignalSource) FiredCount(sig os.Signal) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	count := 0
	for _, s := range f.fired {
		if s == sig {
			count++
		}
	}
	return count
}
