package core

import ts "github.com/tree-sitter/go-tree-sitter"

const (
	FUNC_PRINT = "print"
)

var (
	EMPTY []RslValue
)

func (i *Interpreter) callFunction(
	callNode *ts.Node,
	funcNameNode *ts.Node,
	argValues []RslValue,
	numExpectedOutputs int,
) []RslValue {
	funcName := i.sd.Src[funcNameNode.StartByte():funcNameNode.EndByte()]
	switch funcName {
	case FUNC_PRINT:
		i.assertExpectedNumOutputs(funcNameNode, numExpectedOutputs, 0)
		RP.Print(createPrintStr(argValues))
		return EMPTY
	default:
		i.errorf(funcNameNode, "Unknown function: %s", funcName)
		panic(UNREACHABLE)
	}
}

func createPrintStr(values []RslValue) string {
	if len(values) == 0 {
		return "\n"
	}

	output := ""
	for _, v := range values {
		if str, ok := v.Val.(RslString); ok {
			// explicit handling for string so we don't print surrounding quotes when it's standalone
			output += str.String() + " "
		} else {
			output += ToPrintable(v.Val) + " "
		}
	}
	output = output[:len(output)-1] // remove last space
	output = output + "\n"
	return output
}
