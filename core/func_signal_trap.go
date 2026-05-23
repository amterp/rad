package core

import (
	"github.com/amterp/rad/rts/rl"
)

// FuncSignalTrap registers a Rad function to run when one of the named
// signals is delivered to the process. The handler is invoked with a single
// argument: a map containing the signal name and the conventional exit code.
//
// Replaces any previously-registered handler for the same signal. There is
// no current way to restore the platform default - once a signal is trapped,
// it stays trapped for the lifetime of the interpreter.
//
// After the handler returns, execution always continues - the handler must
// explicitly call exit() if it wants to terminate. This matches Bash, Ruby,
// Python, and Node, which all converge on always-continue. The user pays a
// small ergonomic cost (every SIGINT handler needs `exit(ctx.exit_code)`)
// in exchange for a predictable single-rule mental model.
var FuncSignalTrap = BuiltInFunc{
	Name: FUNC_SIGNAL_TRAP,
	Execute: func(f FuncInvocation) RadValue {
		names, err := extractSignalNames(f, "_signal", FUNC_SIGNAL_TRAP)
		if err != nil {
			return f.Return(err)
		}
		handler := f.GetFn("_handler")
		for _, name := range names {
			if terr := f.i.signals.Trap(name, handler); terr != nil {
				return f.Return(NewErrorStrf("%v", terr).SetCode(rl.ErrGenericRuntime))
			}
		}
		return VOID_SENTINEL
	},
}
