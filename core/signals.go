package core

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
)

// SignalSource is the seam between the SignalManager and the OS signal
// machinery. The production implementation wraps os/signal directly; tests
// inject a fake whose Fire method synchronously delivers a signal to the
// channel without involving the OS (so tests do not race against the real
// process signal queue).
type SignalSource interface {
	Notify(ch chan<- os.Signal, sigs ...os.Signal)
	Stop(ch chan<- os.Signal)
	Ignore(sigs ...os.Signal)
}

// realSignalSource is the production SignalSource. It delegates to the
// stdlib os/signal package. Stateless; safe to instantiate as a value.
type realSignalSource struct{}

func (realSignalSource) Notify(ch chan<- os.Signal, sigs ...os.Signal) {
	signal.Notify(ch, sigs...)
}

func (realSignalSource) Stop(ch chan<- os.Signal) {
	signal.Stop(ch)
}

func (realSignalSource) Ignore(sigs ...os.Signal) {
	signal.Ignore(sigs...)
}

// SignalManager owns signal-handling state for an Interpreter.
//
// It coordinates between the OS signal source (or a test fake) and the
// interpreter's per-statement checkpoint. The OS-facing side runs in a
// dedicated goroutine that drains signals from notifyCh and records them in
// pendingSig; the interpreter side reads pendingSig at safe points between
// statements and dispatches to the registered Rad handler.
//
// Blocking operations (sleep, shell exec, HTTP) consult Ctx() so they can
// wake promptly when a signal arrives - otherwise the handler would not run
// until the blocking call returned on its own.
type SignalManager struct {
	mu sync.Mutex

	// handlers maps a lowercase signal name (e.g. "sigint") to the Rad
	// function that should be invoked when that signal fires. A signal
	// without a handler runs the default action (terminate via
	// RExit.Exit(128+sig) for terminating signals; no-op otherwise).
	handlers map[string]RadFn

	// pendingSigs is the ordered list of signals delivered since the last
	// checkpoint drained them. We dedupe by name so a burst of the same
	// signal coalesces into one handler invocation, but different signals
	// queue separately (so a SIGINT and a SIGTERM arriving back-to-back
	// both get dispatched). Guarded by mu.
	pendingSigs []string

	// inHandlerOf names the signal whose handler is currently executing,
	// or nil. Used to detect double-Ctrl+C (a second SIGINT while a SIGINT
	// handler is in progress triggers an immediate os.Exit(130)).
	inHandlerOf atomic.Pointer[string]

	// ctx is canceled when a signal arrives. Blocking ops select on ctx.Done()
	// so they can return early. The context is recreated after each signal
	// dispatch so subsequent blocking ops are not stuck in the canceled state.
	ctxMu  sync.Mutex
	ctx    context.Context
	cancel context.CancelFunc

	// notifyCh receives signals from the SignalSource (real OS or fake).
	notifyCh chan os.Signal

	// done is closed by Stop to terminate the dispatch goroutine cleanly.
	done chan struct{}

	// started gates the background dispatch goroutine so we start it at most once.
	started bool

	// dispatching is set while a Checkpoint is actively draining the queue
	// and invoking handlers. A nested Checkpoint (e.g. one inside runBlock
	// of the handler's body) bails out instead of recursively popping the
	// queue - the outer Checkpoint will continue draining after the current
	// handler returns. This keeps handler invocations in queue order
	// rather than depth-first.
	dispatching atomic.Bool
}

// NewSignalManager constructs an inactive SignalManager. The dispatch goroutine
// is started lazily by Start (called from Interpreter.Run).
func NewSignalManager() *SignalManager {
	sm := &SignalManager{
		handlers: make(map[string]RadFn),
		notifyCh: make(chan os.Signal, 4),
		done:     make(chan struct{}),
	}
	sm.resetCtx()
	return sm
}

// enqueueSignal appends a signal name to the pending queue. Duplicates are
// dropped: if the same signal is already pending, the second delivery
// coalesces (a burst of SIGINT becomes one handler invocation). Returns
// true if the queue already contained name BEFORE this call - the caller
// uses this to detect a duplicate-in-flight situation (e.g. double-Ctrl+C
// even when the main goroutine is wedged in a blocking op).
func (sm *SignalManager) enqueueSignal(name string) (alreadyPending bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	for _, p := range sm.pendingSigs {
		if p == name {
			return true
		}
	}
	sm.pendingSigs = append(sm.pendingSigs, name)
	return false
}

// resetCtx creates a fresh cancelable context. Called at construction and
// after each signal dispatch so blocking ops started after a handler returns
// are not stuck on the previously-canceled context.
func (sm *SignalManager) resetCtx() {
	sm.ctxMu.Lock()
	defer sm.ctxMu.Unlock()
	sm.ctx, sm.cancel = context.WithCancel(context.Background())
}

// Ctx returns the current signal-cancellation context. Blocking operations
// should pass this to select{} or http.NewRequestWithContext so they can wake
// up promptly when a signal arrives.
func (sm *SignalManager) Ctx() context.Context {
	sm.ctxMu.Lock()
	defer sm.ctxMu.Unlock()
	return sm.ctx
}

// cancelCtx cancels the current signal-cancellation context. Called by the
// dispatch goroutine on signal arrival.
func (sm *SignalManager) cancelCtx() {
	sm.ctxMu.Lock()
	defer sm.ctxMu.Unlock()
	if sm.cancel != nil {
		sm.cancel()
	}
}

// Trap registers fn as the handler for the named signal (e.g. "sigint").
// Replacing an existing handler is allowed; the previous handler is dropped.
// Returns an error if the signal name is not supported on the current
// platform.
func (sm *SignalManager) Trap(name string, fn RadFn) error {
	osSig, _, err := resolveSignalName(name)
	if err != nil {
		return err
	}

	sm.mu.Lock()
	sm.handlers[name] = fn
	sm.mu.Unlock()

	// Subscribe via the injectable source so tests can intercept.
	RSignal.Notify(sm.notifyCh, osSig)
	return nil
}

// Ignore installs OS-level SIG_IGN for the named signal. Useful for SIGPIPE
// when the script pipes its output (e.g. `script | head`) and should not
// terminate when the downstream consumer closes the pipe.
func (sm *SignalManager) Ignore(name string) error {
	osSig, _, err := resolveSignalName(name)
	if err != nil {
		return err
	}
	RSignal.Ignore(osSig)
	return nil
}

// Start launches the background dispatch goroutine and subscribes to the
// signals Rad handles by default (SIGINT, SIGTERM). Idempotent.
//
// Other signals are subscribed lazily when the script calls signal_trap or
// signal_ignore. This keeps the default behavior minimal: Rad takes over
// SIGINT/SIGTERM so it can route them through RExit.Exit (and thereby run
// defer blocks); everything else inherits Go's default behavior until the
// script opts in.
func (sm *SignalManager) Start() {
	sm.mu.Lock()
	if sm.started {
		sm.mu.Unlock()
		return
	}
	sm.started = true
	sm.mu.Unlock()

	// Subscribe to the always-handled defaults.
	for _, name := range []string{sigNameSigint, sigNameSigterm} {
		if osSig, _, err := resolveSignalName(name); err == nil {
			RSignal.Notify(sm.notifyCh, osSig)
		}
		// On Windows, only sigint/sigterm resolve; on unix, both resolve.
		// Either way: silently skip unsupported ones at startup.
	}

	go sm.dispatchLoop()
}

// dispatchLoop is the body of the background goroutine. It reads signals
// from notifyCh and routes them: a second SIGINT while one is already in
// flight force-exits the process; otherwise the signal name is queued for
// the interpreter's checkpoint to drain, and the cancel context is fired
// so blocking ops wake up. The actual handler invocation happens later, in
// the interpreter's checkpoint, on the main script goroutine.
func (sm *SignalManager) dispatchLoop() {
	for {
		select {
		case <-sm.done:
			return
		case sig, ok := <-sm.notifyCh:
			if !ok {
				return
			}
			sm.handleSignal(sig)
		}
	}
}

// handleSignal processes one delivery from notifyCh.
func (sm *SignalManager) handleSignal(sig os.Signal) {
	name := signalToName(sig)
	if name == "" {
		// Unknown signal - shouldn't happen since we only subscribe to
		// signals we explicitly resolved, but be defensive.
		return
	}

	// Double-SIGINT force-exit: if a SIGINT is already in flight - either
	// a SIGINT handler is currently running, OR a previous SIGINT is queued
	// but the main goroutine hasn't reached the next checkpoint yet (e.g.
	// stuck in a subprocess wait or other blocking op that won't surrender) -
	// the second SIGINT is the user's escape hatch and exits immediately.
	// Defers do not run; this is a force-exit contract.
	if name == sigNameSigint {
		if cur := sm.inHandlerOf.Load(); cur != nil && *cur == sigNameSigint {
			os.Exit(130)
		}
		if sm.enqueueSignal(name) {
			os.Exit(130)
		}
	} else {
		sm.enqueueSignal(name)
	}

	// Wake up any blocking ops so the checkpoint fires promptly.
	sm.cancelCtx()
}

// DrainPending returns the next pending signal (if any) and removes it from
// the queue. Called by the interpreter's checkpoint between statements.
func (sm *SignalManager) DrainPending() (string, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	if len(sm.pendingSigs) == 0 {
		return "", false
	}
	name := sm.pendingSigs[0]
	sm.pendingSigs = sm.pendingSigs[1:]
	return name, true
}

// Handler returns the registered handler for the named signal. The second
// return is false if no handler is registered.
func (sm *SignalManager) Handler(name string) (RadFn, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	fn, ok := sm.handlers[name]
	return fn, ok
}

// MarkInHandler / ClearInHandler bracket handler execution for double-Ctrl+C
// detection.
func (sm *SignalManager) MarkInHandler(name string) {
	nameCopy := name
	sm.inHandlerOf.Store(&nameCopy)
}

func (sm *SignalManager) ClearInHandler() {
	sm.inHandlerOf.Store(nil)
}

// RefreshCtx is called after a handler returns (or a default-terminate
// decision is made and forgone) so the next blocking op gets a fresh
// (non-canceled) context.
func (sm *SignalManager) RefreshCtx() {
	sm.resetCtx()
}

// Stop tears down the dispatch goroutine and OS subscriptions. Safe to call
// multiple times. Called from Interpreter.Run's defer so embedded/test usage
// doesn't accumulate leaked goroutines across interpreter lifetimes.
func (sm *SignalManager) Stop() {
	sm.mu.Lock()
	if !sm.started {
		sm.mu.Unlock()
		return
	}
	sm.started = false
	sm.mu.Unlock()

	RSignal.Stop(sm.notifyCh)
	close(sm.done)
}

// NotifyCh returns the channel that subscribed signals are delivered to.
// Tests use this with FakeSignalSource to inject signals deterministically
// (write to this channel directly, then wait for DrainPending to observe
// the result). Not part of the public Rad API.
func (sm *SignalManager) NotifyCh() chan<- os.Signal {
	return sm.notifyCh
}

// signalToName converts an os.Signal back to its lowercase Rad name (e.g.
// syscall.SIGINT -> "sigint"). Returns "" if not recognized.
func signalToName(sig os.Signal) string {
	for name, entry := range supportedSignals() {
		if entry.osSignal == sig {
			return name
		}
	}
	return ""
}

// signalEntry describes a supported signal on the current platform.
type signalEntry struct {
	osSignal os.Signal
	sigNum   int // POSIX signal number, used for 128+sig exit codes
}

// ExitCodeFor returns the conventional exit code (128 + signal number) for
// the named signal. Returns 0 if the signal is not supported.
func ExitCodeFor(name string) int {
	if entry, ok := supportedSignals()[name]; ok {
		return 128 + entry.sigNum
	}
	return 0
}

// IsTerminatingSignal reports whether the named signal's default action is to
// terminate the process. Used by the checkpoint when no handler is registered:
// terminating signals cause RExit.Exit(128+sig), non-terminating ones (only
// SIGWINCH today) are no-ops.
func IsTerminatingSignal(name string) bool {
	switch name {
	case sigNameSigwinch:
		return false
	}
	return true
}

// Checkpoint is called by the interpreter between statements. If a signal is
// pending it dispatches according to the current policy:
//
//   - If a Rad handler is registered for the signal, invoke it with a single
//     map argument: { signal, exit_code }. After it returns, always continue
//     execution (Bash/Python/Ruby/Node semantics). The handler must call
//     exit() explicitly if it wants to terminate.
//
//   - If no handler is registered and the signal is terminating, route
//     through RExit.Exit(128+sig) so defer blocks run and the exit code
//     reflects the signal.
//
//   - If no handler and the signal is non-terminating (SIGWINCH), drop it.
//
// Checkpoint runs on the main interpreter goroutine, so it can safely touch
// the interpreter env / call deferred blocks / invoke Rad handlers.
//
// Drains the entire pending queue in FIFO order before returning. The
// dispatching guard ensures nested checkpoints (e.g. those inside runBlock
// while a handler body is executing) become no-ops, so handler invocations
// stay in queue order rather than interleaving depth-first.
func (i *Interpreter) Checkpoint() {
	if !i.signals.dispatching.CompareAndSwap(false, true) {
		return
	}
	defer i.signals.dispatching.Store(false)

	for {
		name, ok := i.signals.DrainPending()
		if !ok {
			return
		}
		i.dispatchSignal(name)
	}
}

// dispatchSignal handles one drained signal: invokes the Rad handler if one
// is registered, otherwise routes terminating signals through RExit.Exit so
// defers run. Always-continue semantics: even if a handler is registered,
// control returns to the caller after it completes.
func (i *Interpreter) dispatchSignal(name string) {
	// Refresh the context BEFORE dispatching so the handler's own blocking
	// ops (sleep, shell, HTTP) see a non-canceled context. If a new signal
	// arrives during the handler, the dispatch goroutine will cancel this
	// fresh context and the queue drain in the outer Checkpoint loop picks
	// it up on the next iteration.
	i.signals.RefreshCtx()

	if handler, hasHandler := i.signals.Handler(name); hasHandler {
		// Mark in-handler so a second SIGINT during a SIGINT handler triggers
		// the double-Ctrl+C force-exit path in dispatchLoop. The mark is
		// signal-scoped: a SIGTERM arriving during a SIGINT handler is
		// queued normally; it would only collide with itself.
		i.signals.MarkInHandler(name)
		defer i.signals.ClearInHandler()

		ctxMap := NewRadMap()
		ctxMap.SetPrimitiveStr("signal", name)
		ctxMap.SetPrimitiveInt("exit_code", ExitCodeFor(name))
		ctxVal := newRadValue(i, nil, ctxMap)

		invocation := NewFnInvocation(
			i,
			nil, // signal handlers have no syntactic call site
			handler.Name(),
			NewPosArgs(NewPosArg(nil, ctxVal)),
			NO_NAMED_ARGS_INPUT,
			handler.IsBuiltIn(),
		)
		_ = handler.Execute(invocation)
		// Always continue after the handler returns. If the handler wanted
		// to exit, it called exit() itself, which doesn't return.
		return
	}

	if IsTerminatingSignal(name) {
		RExit.Exit(ExitCodeFor(name))
	}
}

// Stable lowercase names so we can compare against string keys without
// import cycles or stringly-typed misspellings.
const (
	sigNameSigint   = "sigint"
	sigNameSigterm  = "sigterm"
	sigNameSighup   = "sighup"
	sigNameSigusr1  = "sigusr1"
	sigNameSigusr2  = "sigusr2"
	sigNameSigpipe  = "sigpipe"
	sigNameSigwinch = "sigwinch"
)

// allKnownSignalNames is the canonical set of signal names Rad ever exposes.
// The current platform may support a subset (Windows in particular only
// supports sigint/sigterm). Lets us distinguish "unknown signal name" (always
// an error) from "supported on some platforms but not this one" (a clearer
// diagnostic).
var allKnownSignalNames = map[string]struct{}{
	sigNameSigint:   {},
	sigNameSigterm:  {},
	sigNameSighup:   {},
	sigNameSigusr1:  {},
	sigNameSigusr2:  {},
	sigNameSigpipe:  {},
	sigNameSigwinch: {},
}

func isKnownSignalName(name string) bool {
	_, ok := allKnownSignalNames[name]
	return ok
}

// errUnsupportedSignal is returned when the script asks for a signal that the
// current platform does not support (e.g. SIGUSR1 on Windows). Lists what IS
// supported so the user has a path forward.
func errUnsupportedSignal(name string) error {
	return fmt.Errorf("signal %q is not supported on this platform (supported here: %s)", name, strings.Join(supportedSignalsList(), ", "))
}

// errUnknownSignal is returned when the script asks for a signal name that is
// not in the supported set at all. If the name matches a known signal once
// case-folded, the message includes a "did you mean" hint - this catches the
// common SIGINT-vs-sigint mistake when the value escapes the signature's
// static enum check (e.g. computed at runtime).
func errUnknownSignal(name string) error {
	if lower := strings.ToLower(name); lower != name && isKnownSignalName(lower) {
		return fmt.Errorf("unknown signal %q - signal names are lowercase in Rad; did you mean %q?", name, lower)
	}
	return fmt.Errorf("unknown signal %q (valid: sigint, sigterm, sighup, sigusr1, sigusr2, sigpipe, sigwinch)", name)
}

// supportedSignalsList returns the platform's supported signal names in the
// canonical declaration order (sigint, sigterm, ...) so error messages read
// naturally rather than alphabetically.
func supportedSignalsList() []string {
	m := supportedSignals()
	ordered := []string{
		sigNameSigint, sigNameSigterm, sigNameSighup,
		sigNameSigusr1, sigNameSigusr2, sigNameSigpipe, sigNameSigwinch,
	}
	out := make([]string, 0, len(m))
	for _, n := range ordered {
		if _, ok := m[n]; ok {
			out = append(out, n)
		}
	}
	return out
}
