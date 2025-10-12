package core

import (
	"fmt"
)

var FuncExit = BuiltInFunc{
	Name: FUNC_EXIT,
	Execute: func(f FuncInvocation) RadValue {
		err := f.GetIntAllowingBool("_code")
		exit(f.i, err)
		return VOID_SENTINEL
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
	RExit.Exit(int(errorCode))
}
