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
