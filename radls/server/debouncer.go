package server

import (
	"sync"
	"time"
)

// Debouncer batches per-key triggers into a single delayed callback.
//
// Designed for the LSP diagnostics-publishing pattern: every didChange
// fires Trigger(uri, publishFn); the callback runs once the URI has
// been quiet for `delay`. If another Trigger arrives before the timer
// fires, the timer is reset, so a burst of rapid keystrokes results in
// exactly one publish at the end - not one per stroke.
//
// Analysis still runs synchronously on every didChange (so a hover or
// goto-def landing mid-burst still sees the latest snapshot). What we
// debounce is the wire publish - the noisy thing that creates "errors
// flicker as I type" UX, not the freshness of internal state.
//
// Concurrency: Trigger is safe to call from multiple goroutines; the
// callback runs in its own goroutine (spawned by time.AfterFunc).
// Stop() cancels all pending timers and is safe to call exactly once
// at server shutdown.
type Debouncer struct {
	delay time.Duration

	mu      sync.Mutex
	timers  map[string]*time.Timer
	stopped bool
}

// NewDebouncer returns a Debouncer with the given idle window. A
// typical value for LSP diagnostics is 150-250ms; long enough that a
// burst of typing doesn't publish per keystroke, short enough that
// the user sees errors when they pause.
func NewDebouncer(delay time.Duration) *Debouncer {
	return &Debouncer{
		delay:  delay,
		timers: make(map[string]*time.Timer),
	}
}

// Trigger schedules `fn` to run after the configured delay. If
// another Trigger lands on the same key first, the existing timer is
// stopped and the new fn replaces it - the most recent fn wins. This
// matters because the publish closure usually wants to grab the
// latest snapshot at fire time, not the one at trigger time, so it's
// fine that earlier fns are dropped.
//
// A zero delay short-circuits to a synchronous call: fn runs inline
// on the caller's goroutine. Useful in tests that want deterministic
// publish behavior without racing a goroutine scheduler.
//
// If the Debouncer has been Stopped, Trigger is a no-op.
func (d *Debouncer) Trigger(key string, fn func()) {
	if d.delay == 0 {
		// Synchronous fast path. Skip the timer + goroutine machinery
		// entirely so tests don't need to sleep or flush at shutdown.
		d.mu.Lock()
		stopped := d.stopped
		d.mu.Unlock()
		if stopped {
			return
		}
		fn()
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if d.stopped {
		return
	}
	if existing, ok := d.timers[key]; ok {
		existing.Stop()
	}
	d.timers[key] = time.AfterFunc(d.delay, func() {
		fn()
		d.mu.Lock()
		delete(d.timers, key)
		d.mu.Unlock()
	})
}

// Stop cancels all pending timers. Idempotent; after Stop, subsequent
// Trigger calls are no-ops.
func (d *Debouncer) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.stopped = true
	for _, t := range d.timers {
		t.Stop()
	}
	d.timers = nil
}
