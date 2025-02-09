package core

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncExit = Func{
	Name:             FUNC_EXIT,
	ReturnValues:     ZERO_RETURN_VALS,
	RequiredArgCount: 0,
	ArgTypes:         [][]RslTypeEnum{{RslIntT}},
	NamedArgs:        NO_NAMED_ARGS,
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
		if len(args) == 0 {
			exit(i, 0)
		} else {
			arg := args[0]
			exit(i, arg.value.RequireInt(i, arg.node))
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
