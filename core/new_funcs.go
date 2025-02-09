package core

import ts "github.com/tree-sitter/go-tree-sitter"

const (
	FUNC_PRINT = "print"
	FUNC_LEN   = "len"
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
		i.assertExpectedNumOutputs(callNode, numExpectedOutputs, 0)
		RP.Print(createPrintStr(argValues))
		return EMPTY
	case FUNC_LEN:
		i.assertExpectedNumOutputs(callNode, numExpectedOutputs, 1)
		if len(argValues) != 1 {
			i.errorf(callNode, "%s() takes exactly one argument", FUNC_LEN)
		}
		switch v := argValues[0].Val.(type) {
		case RslString:
			return newRslValues(i, callNode, v.Len())
		case *RslList:
			return newRslValues(i, callNode, v.Len())
		case *RslMap:
			return newRslValues(i, callNode, v.Len())
		default:
			i.errorf(callNode, "%s() takes a string or collection", FUNC_LEN)
			panic(UNREACHABLE)
		}
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
