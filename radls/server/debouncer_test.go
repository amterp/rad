package server

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestDebouncerCollapsesBurst(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)
	defer d.Stop()

	var calls atomic.Int32
	for i := 0; i < 10; i++ {
		d.Trigger("k", func() { calls.Add(1) })
	}
	// Wait long enough for any timer to fire.
	time.Sleep(150 * time.Millisecond)
	if got := calls.Load(); got != 1 {
		t.Errorf("burst of 10 triggers: got %d calls, want 1", got)
	}
}

func TestDebouncerLastFnWins(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)
	defer d.Stop()

	var observed atomic.Int32
	d.Trigger("k", func() { observed.Store(1) })
	d.Trigger("k", func() { observed.Store(2) })
	d.Trigger("k", func() { observed.Store(3) })
	time.Sleep(150 * time.Millisecond)
	if got := observed.Load(); got != 3 {
		t.Errorf("expected last fn (3) to win, got %d", got)
	}
}

func TestDebouncerPerKeyIsolated(t *testing.T) {
	d := NewDebouncer(50 * time.Millisecond)
	defer d.Stop()

	var a, b atomic.Int32
	d.Trigger("a", func() { a.Add(1) })
	d.Trigger("b", func() { b.Add(1) })
	time.Sleep(150 * time.Millisecond)
	if a.Load() != 1 || b.Load() != 1 {
		t.Errorf("expected one call per key; got a=%d b=%d", a.Load(), b.Load())
	}
}

func TestDebouncerStopCancelsPending(t *testing.T) {
	d := NewDebouncer(100 * time.Millisecond)

	var calls atomic.Int32
	d.Trigger("k", func() { calls.Add(1) })
	d.Stop()
	time.Sleep(200 * time.Millisecond)
	if got := calls.Load(); got != 0 {
		t.Errorf("Stop should cancel pending timer; got %d calls", got)
	}

	// Trigger after stop is a no-op.
	d.Trigger("k", func() { calls.Add(1) })
	time.Sleep(200 * time.Millisecond)
	if got := calls.Load(); got != 0 {
		t.Errorf("Trigger after Stop should no-op; got %d calls", got)
	}
}

func TestDebouncerConcurrentTriggers(t *testing.T) {
	d := NewDebouncer(20 * time.Millisecond)
	defer d.Stop()

	var wg sync.WaitGroup
	var calls atomic.Int32
	const n = 200
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			d.Trigger("k", func() { calls.Add(1) })
		}()
	}
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	got := calls.Load()
	// All concurrent triggers should debounce to exactly 1 call.
	if got != 1 {
		t.Errorf("concurrent triggers: got %d calls, want 1", got)
	}
}

// TestDebouncerSupersededCallbackDoesNotDeleteSuccessor regression
// tests the race the bug-hunter identified:
//
//  1. Timer T1 fires for key K. Its callback is in fn(); it has not
//     yet taken d.mu to delete itself from the map.
//  2. A new Trigger arrives for K. Stop() returns false (T1 already
//     fired). Trigger installs T2 in d.timers[K].
//  3. T1's callback finally takes d.mu and does delete(d.timers, K).
//     Without the generation check, this deletes T2's entry.
//  4. A third Trigger for K sees no existing entry and installs T3.
//     T2 and T3 both fire - duplicate publish, contract violated.
//
// We engineer the race by holding a slow callback (so it's clearly
// still in fn() when the next Trigger lands) and verify that after
// a burst of Trigger calls, exactly one further fn fires.
func TestDebouncerSupersededCallbackDoesNotDeleteSuccessor(t *testing.T) {
	d := NewDebouncer(10 * time.Millisecond)
	defer d.Stop()

	// Calls counts how many fns from the "next" burst actually fire.
	var nextBurstCalls atomic.Int32

	// Block the first callback in flight so we have a deterministic
	// "fired-but-running" window when the next Trigger arrives.
	firstFiring := make(chan struct{})
	releaseFirst := make(chan struct{})
	d.Trigger("k", func() {
		close(firstFiring)
		<-releaseFirst
	})

	// Wait for the first callback to enter fn(). At this point its
	// timer has fired and Stop() would return false.
	<-firstFiring

	// Now schedule a new Trigger. With the bug, when the first
	// callback eventually finishes, it would delete this new
	// timer's map entry.
	d.Trigger("k", func() {
		nextBurstCalls.Add(1)
	})

	// Release the first callback. With the fix, its delete is gated
	// on its generation still matching - it won't.
	close(releaseFirst)

	// Wait long enough for the second timer to fire normally.
	time.Sleep(200 * time.Millisecond)

	// And now schedule one more. If the second timer's map entry was
	// erroneously deleted by the first callback, this third Trigger
	// would install a fresh timer rather than coalescing into the
	// already-fired-or-pending second - resulting in 2 fires of
	// "nextBurst" fns total.
	d.Trigger("k", func() {
		nextBurstCalls.Add(1)
	})

	// Wait for the third timer to fire.
	time.Sleep(200 * time.Millisecond)

	if got := nextBurstCalls.Load(); got != 2 {
		t.Errorf("expected 2 next-burst fires (one for each Trigger after the slow first); got %d", got)
	}
}

func TestDebouncerZeroDelaySynchronous(t *testing.T) {
	d := NewDebouncer(0)
	defer d.Stop()

	// With zero delay, Trigger should call fn inline before returning.
	// No sleep needed - if it didn't, the assertion fails immediately.
	var called bool
	d.Trigger("k", func() { called = true })
	if !called {
		t.Error("zero-delay Trigger should call fn synchronously")
	}
}
