package core

// ExitReason says why execution is ending, which a REPL needs in order to decide
// what to do about it: an explicit exit() ends the session, a rendered error just
// returns you to the prompt.
type ExitReason int

const (
	ExitExplicit ExitReason = iota // exit(), or the script simply finishing
	ExitError                      // a diagnostic has already been rendered
	ExitSignal                     // a terminating signal with no handler
)

// RadAbort unwinds execution instead of ending the process. It is panicked by
// the exit handler when unwinding is enabled, and recovered by whoever owns the
// unit of execution - in practice, the REPL's per-turn boundary.
//
// A panic rather than a return is what makes this safe to bolt onto an
// interpreter written for a one-shot process: only some of the ~50 error sites
// are followed by panic(UNREACHABLE), and the rest simply return, so an abort
// that returned would resume evaluation on state the error just invalidated.
type RadAbort struct {
	Code   int
	Reason ExitReason
	Signal string
}

func (a *RadAbort) Error() string { return "rad: execution aborted" }

type RadExitHandler struct {
	delegate             func(int)
	preExitCallbacks     []func()
	executeDeferredStmts func(code int)

	// unwind makes Exit panic a *RadAbort instead of calling the delegate, so a
	// long-lived session survives an error that would end a script.
	unwind bool

	// stateful for when exiting
	exiting        bool
	codeToExitWith int
}

func NewExitHandler(delegate func(int)) *RadExitHandler {
	return &RadExitHandler{
		delegate:             delegate,
		executeDeferredStmts: func(code int) {},
	}
}

// SetUnwinding switches between ending the process and unwinding to whoever is
// hosting execution. Scripts leave it off; the REPL turns it on for the session.
func (h *RadExitHandler) SetUnwinding(unwind bool) { h.unwind = unwind }

// ResetExiting clears the one-shot latch so the next unit of execution starts
// fresh. Without it the first error in a session would leave every later exit
// short-circuiting past deferred statements and pre-exit callbacks.
func (h *RadExitHandler) ResetExiting() {
	h.exiting = false
	h.codeToExitWith = 0
}

func (h *RadExitHandler) Exit(code int) {
	h.ExitWith(code, ExitExplicit, "")
}

// ExitWith is Exit with the reason attached, for the callers that know it.
func (h *RadExitHandler) ExitWith(code int, reason ExitReason, signal string) {
	RP.RadDebugf("Exit handler called with code %d", code)
	if h.exiting {
		RP.RadDebugf("Already exiting")
		// we're already exiting. if we're here again, it's probably because one of the deferred
		// statements is calling exit again (perhaps because it failed). we should keep running
		// all the deferred statements, however, and *then* exit.
		// we escalate the exit code to error if we're now being invoked with an error code.
		if h.codeToExitWith == 0 {
			h.codeToExitWith = code
		}
		return
	}

	h.exiting = true
	h.codeToExitWith = code
	// todo gets executed *after* any error is printed (if error), should delay error print until after (i think?)
	RP.RadDebugf("Executing deferred statements")
	h.executeDeferredStmts(h.codeToExitWith)
	RP.RadDebugf("Executing pre exit callbacks")
	for _, cb := range h.preExitCallbacks {
		RP.RadDebugf("Executing next pre exit callback")
		cb()
	}
	if h.unwind {
		RP.RadDebugf("Unwinding with code %d instead of exiting", h.codeToExitWith)
		panic(&RadAbort{Code: h.codeToExitWith, Reason: reason, Signal: signal})
	}
	RP.RadDebugf("Executing exit delegate with code %d", h.codeToExitWith)
	h.delegate(h.codeToExitWith)
}

func (h *RadExitHandler) AddPreExitCallback(cb func()) {
	h.preExitCallbacks = append(h.preExitCallbacks, cb)
}

func (h *RadExitHandler) SetExecuteDeferredStmtsFunc(f func(code int)) {
	h.executeDeferredStmts = f
}
