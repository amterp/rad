package core

import ts "github.com/tree-sitter/go-tree-sitter"

var FuncPrint = Func{
	Name:             FUNC_PRINT,
	ReturnValues:     ZERO_RETURN_VALS,
	RequiredArgCount: 0,
	ArgTypes:         [][]RslTypeEnum{{}},
	NamedArgs: map[string][]RslTypeEnum{
		"reverse": {RslBoolT},
	},
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, _ map[string]namedArg) []RslValue {
		RP.Print(createPrintStr(args))
		return EMPTY
	},
}

func createPrintStr(values []positionalArg) string {
	if len(values) == 0 {
		return "\n"
	}

	output := ""
	for _, v := range values {
		if v.value.Type() == RslStringT {
			// explicit handling for string so we don't print surrounding quotes when it's standalone
			output += ToPrintableQuoteStr(v.value.Val, false)
		} else {
			output += ToPrintableQuoteStr(v.value.Val, true)
		}
		output += " "
	}
	output = output[:len(output)-1] // remove last space
	output = output + "\n"
	return output
}
