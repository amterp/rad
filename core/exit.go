package core

type RadExitHandler struct {
	delegate             func(int)
	preExitCallbacks     []func()
	executeDeferredStmts func(code int)

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

func (h *RadExitHandler) Exit(code int) {
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
	RP.RadDebugf("Executing exit delegate with code %d", h.codeToExitWith)
	h.delegate(h.codeToExitWith)
}

func (h *RadExitHandler) AddPreExitCallback(cb func()) {
	h.preExitCallbacks = append(h.preExitCallbacks, cb)
}

func (h *RadExitHandler) SetExecuteDeferredStmtsFunc(f func(code int)) {
	h.executeDeferredStmts = f
}
