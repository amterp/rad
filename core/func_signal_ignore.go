package core

import (
	"github.com/amterp/rad/rts/rl"
)

// FuncSignalIgnore installs OS-level SIG_IGN for one or more signals. Most
// useful for SIGPIPE: a script that pipes its output (e.g. `script | head`)
// will otherwise crash when the downstream consumer closes the pipe.
//
// Distinct from `signal_trap(sig, fn(){})`: a no-op handler still wakes the
// process and runs Rad-side dispatch on every delivery, whereas SIG_IGN tells
// the kernel to drop the signal before it reaches us. For SIGPIPE in
// particular this also affects subprocess inheritance.
var FuncSignalIgnore = BuiltInFunc{
	Name: FUNC_SIGNAL_IGNORE,
	Execute: func(f FuncInvocation) RadValue {
		names, err := extractSignalNames(f, "_signal", FUNC_SIGNAL_IGNORE)
		if err != nil {
			return f.Return(err)
		}
		for _, name := range names {
			if ierr := f.i.signals.Ignore(name); ierr != nil {
				return f.Return(NewErrorStrf("%v", ierr).SetCode(rl.ErrGenericRuntime))
			}
		}
		return VOID_SENTINEL
	},
}

// extractSignalNames pulls a signal name or list of signal names from the
// given arg. Shared between signal_ignore and signal_trap.
func extractSignalNames(f FuncInvocation, argName, funcName string) ([]string, *RadError) {
	arg := f.GetArg(argName)
	switch coerced := arg.Val.(type) {
	case RadString:
		return []string{coerced.Plain()}, nil
	case *RadList:
		names := make([]string, 0, len(coerced.Values))
		for _, item := range coerced.Values {
			names = append(names, item.RequireStr(f.i, f.callNode).Plain())
		}
		return names, nil
	default:
		bugIncorrectTypes(funcName)
		panic(UNREACHABLE)
	}
}
