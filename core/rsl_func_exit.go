package core

import (
	"fmt"
)

var FuncExit = Func{
	Name:           FUNC_EXIT,
	ReturnValues:   ZERO_RETURN_VALS,
	MinPosArgCount: 0,
	PosArgTypes:    [][]RslTypeEnum{{RslIntT}},
	NamedArgs:      NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) []RslValue {
		if len(f.args) == 0 {
			exit(f.i, 0)
		} else {
			arg := f.args[0]
			exit(f.i, arg.value.RequireInt(f.i, arg.node))
		}
		return EMPTY
	},
}

func exit(i *Interpreter, errorCode int64) {
	if FlagShell.Value {
		if errorCode == 0 {
			RP.RadDebugf(fmt.Sprintf("Printing shell exports"))
			i.env.PrintShellExports()
		} else {
			// error scenario, we want the shell script to exit, so just print a shell exit to be eval'd
			RP.RadDebugf(fmt.Sprintf("Printing shell exit %d", errorCode))
			RP.PrintForShellEval(fmt.Sprintf("exit %d\n", errorCode))
		}
	}

	RP.RadDebugf("Exiting")
	RExit(int(errorCode))
}
