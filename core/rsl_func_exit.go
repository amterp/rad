package core

import (
	"fmt"
)

func runExit(i *MainInterpreter, function Token, args []interface{}) {
	if len(args) == 0 {
		exit(i, 0)
	} else if len(args) == 1 {
		arg := args[0]
		switch coerced := arg.(type) {
		case int64:
			exit(i, int(coerced))
		default:
			i.error(function, EXIT+"() takes an integer argument")
		}
	} else {
		i.error(function, EXIT+"() takes zero or one argument")
	}
}

func exit(i *MainInterpreter, errorCode int) {
	if shellFlag {
		if errorCode == 0 {
			RP.RadDebug(fmt.Sprintf("Printing shell exports"))
			i.env.PrintShellExports()
		} else {
			// error scenario, we want the shell script to exit, so just print a shell exit to be eval'd
			RP.RadDebug(fmt.Sprintf("Printing shell exit %d", errorCode))
			RP.PrintForShellEval(fmt.Sprintf("exit %d\n", errorCode))
		}
	}

	RP.RadDebug("Exiting")
	RExit(errorCode)
}
