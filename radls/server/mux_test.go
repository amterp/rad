package server

import (
	"bufio"
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"

	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"
)

func init() {
	if log.L == nil {
		log.L = zap.NewNop().Sugar()
	}
}

// TestCancelRequestPropagatesToHandler verifies the load-bearing
// property of Phase 8e: a handler that blocks on ctx.Done() unblocks
// when the matching $/cancelRequest arrives. We call into the Mux
// directly rather than through pipes to keep the test deterministic.
func TestCancelRequestPropagatesToHandler(t *testing.T) {
	m := NewMux(nil, nil)
	defer m.baseCancel()
	m.writer = bufio.NewWriter(discardWriter{})

	started := make(chan struct{})
	finished := make(chan struct{})

	m.AddRequestHandler("test/slow", func(ctx context.Context, params json.RawMessage) (any, error) {
		close(started)
		<-ctx.Done()
		return "cancelled", nil
	})

	reqID := json.RawMessage(`42`)
	params := json.RawMessage(`null`)
	req := lsp.Request{
		Msg:    lsp.Msg{Rpc: "2.0"},
		Id:     &reqID,
		Method: "test/slow",
		Params: &params,
	}

	go func() {
		_ = m.handleRequestResponse(req)
		close(finished)
	}()

	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("handler never started")
	}

	// Fire the cancel as a real notification through Mux. The
	// inflight registry should resolve id 42 to the handler's ctx.
	cancelParams := json.RawMessage(`{"id":42}`)
	cancelNotif := lsp.Notification{
		Msg:    lsp.Msg{Rpc: "2.0"},
		Method: lsp.CANCEL_REQUEST,
		Params: &cancelParams,
	}
	if err := m.handleNotification(cancelNotif); err != nil {
		t.Fatalf("handleNotification: %v", err)
	}

	select {
	case <-finished:
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not unblock after cancellation")
	}
}

// TestCancelUnknownIdIsBenign verifies that $/cancelRequest for an id
// that isn't (or no longer is) in flight just logs and returns - no
// panic, no error, no side effect.
func TestCancelUnknownIdIsBenign(t *testing.T) {
	m := NewMux(nil, nil)
	defer m.baseCancel()
	params := json.RawMessage(`{"id":999}`)
	err := m.handleCancelRequest(context.Background(), params)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestCancelMalformedParamsIsBenign verifies that a malformed cancel
// payload is logged and dropped, not propagated as an error.
func TestCancelMalformedParamsIsBenign(t *testing.T) {
	m := NewMux(nil, nil)
	defer m.baseCancel()
	params := json.RawMessage(`{"not-an-id":42}`)
	err := m.handleCancelRequest(context.Background(), params)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

// TestInflightRegistrationCleanup verifies the in-flight registry
// shrinks back to empty after a request completes. Critical: the
// inflight map is the only thing keeping references to cancel
// closures, so a leak here is a memory + goroutine leak.
func TestInflightRegistrationCleanup(t *testing.T) {
	m := NewMux(nil, nil)
	defer m.baseCancel()

	done := make(chan struct{})
	m.AddRequestHandler("test/quick", func(ctx context.Context, params json.RawMessage) (any, error) {
		close(done)
		return "ok", nil
	})

	// Pump a request through handleRequestResponse directly so we
	// don't need pipes. Use a no-op writer.
	m.writer = bufio.NewWriter(discardWriter{})

	reqID := json.RawMessage(`7`)
	params := json.RawMessage(`null`)
	req := lsp.Request{
		Msg:    lsp.Msg{Rpc: "2.0"},
		Id:     &reqID,
		Method: "test/quick",
		Params: &params,
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = m.handleRequestResponse(req)
	}()
	<-done
	wg.Wait()

	m.inflightMu.Lock()
	n := len(m.inflight)
	m.inflightMu.Unlock()
	if n != 0 {
		t.Errorf("inflight registry should be empty after request returns; got %d entries", n)
	}
}

// TestRegisterInflightCancelsClobberedEntry verifies that if a
// duplicate request id is registered while the prior one is still
// in-flight (a misbehaving client scenario), the old cancel func
// is fired so the prior goroutine isn't stuck waiting on a context
// that nobody ever cancels.
func TestRegisterInflightCancelsClobberedEntry(t *testing.T) {
	m := NewMux(nil, nil)
	defer m.baseCancel()

	// First registration: install a cancel that signals when fired.
	fired := make(chan struct{})
	m.registerInflight("dup", func() { close(fired) })

	// Second registration with the same key: should cancel the first.
	m.registerInflight("dup", func() {})

	select {
	case <-fired:
	case <-time.After(time.Second):
		t.Fatal("clobbered cancel func was never called")
	}

	// Cleanup.
	m.unregisterInflight("dup")
}

// TestRequestIDKeyDistinguishesNumberAndString documents that the JSON
// id `1` (number) and `"1"` (string) are different keys, matching JSON-RPC.
func TestRequestIDKeyDistinguishesNumberAndString(t *testing.T) {
	num := json.RawMessage(`1`)
	str := json.RawMessage(`"1"`)
	if requestIDKey(&num) == requestIDKey(&str) {
		t.Errorf("number and string ids should be distinct keys")
	}
}

// discardWriter is io.Writer that swallows everything. Used so the
// no-pipe tests can drive handleRequestResponse without needing a
// reader on the other end.
type discardWriter struct{}

func (discardWriter) Write(p []byte) (int, error) { return len(p), nil }
