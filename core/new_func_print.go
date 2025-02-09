package core

import (
	"bytes"
	"strings"

	"github.com/nwidger/jsoncolor"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncPrint = Func{
	Name:             FUNC_PRINT,
	ReturnValues:     ZERO_RETURN_VALS,
	RequiredArgCount: 0,
	// TODO BAD!! We need a way to say 'unlimited positional args'
	ArgTypes:  [][]RslTypeEnum{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
	NamedArgs: NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) []RslValue {
		RP.Print(resolvePrintStr(f.args))
		return EMPTY
	},
}

var FuncPPrint = Func{
	Name:             FUNC_PPRINT,
	ReturnValues:     ZERO_RETURN_VALS,
	RequiredArgCount: 0,
	ArgTypes:         [][]RslTypeEnum{{}},
	NamedArgs:        NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) []RslValue {
		if len(f.args) == 0 {
			RP.Print("\n")
		}

		arg := f.args[0]
		jsonStruct := RslToJsonType(arg.value)
		output := prettify(f.i, f.callNode, jsonStruct)
		RP.Print(output)
		return EMPTY
	},
}

var FuncDebug = Func{
	Name:             FUNC_DEBUG,
	ReturnValues:     ZERO_RETURN_VALS,
	RequiredArgCount: 0,
	// TODO BAD!! We need a way to say 'unlimited positional args'
	ArgTypes:  [][]RslTypeEnum{{}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}, {}},
	NamedArgs: NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) []RslValue {
		RP.ScriptDebug(resolvePrintStr(f.args))
		return EMPTY
	},
}

func resolvePrintStr(args []positionalArg) string {
	var sb strings.Builder

	if len(args) == 0 {
		sb.WriteString("\n")
	} else {
		for idx, v := range args {
			if v.value.Type() == RslStringT {
				// explicit handling for string so we don't print surrounding quotes when it's standalone
				sb.WriteString(ToPrintableQuoteStr(v.value.Val, false))
			} else {
				sb.WriteString(ToPrintableQuoteStr(v.value.Val, true))
			}
			if idx < len(args)-1 {
				sb.WriteString(" ")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

func prettify(i *Interpreter, callNode *ts.Node, jsonStruct interface{}) string {
	f := jsoncolor.NewFormatter()
	// todo could add coloring here on formatter

	buf := &bytes.Buffer{}

	enc := jsoncolor.NewEncoderWithFormatter(buf, f)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)

	err := enc.Encode(jsonStruct)

	if err != nil {
		i.errorf(callNode, "Error marshalling JSON: %v", err)
	}

	return buf.String()
}
