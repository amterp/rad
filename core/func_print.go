package core

import (
	"bytes"
	"strings"

	"github.com/amterp/jsoncolor"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncPrint = BuiltInFunc{
	Name:            FUNC_PRINT,
	ReturnValues:    ZERO_RETURN_VALS,
	MinPosArgCount:  0,
	PosArgValidator: NewVarArgSchema([]RslTypeEnum{}),
	NamedArgs: map[string][]RslTypeEnum{
		namedArgEnd: {RslStringT},
		namedArgSep: {RslStringT},
	},
	Execute: func(f FuncInvocationArgs) []RslValue {
		RP.Print(resolvePrintStr(f))
		return EMPTY
	},
}

var FuncPPrint = BuiltInFunc{
	Name:            FUNC_PPRINT,
	ReturnValues:    ZERO_RETURN_VALS,
	MinPosArgCount:  0,
	PosArgValidator: NewEnumerableArgSchema([][]RslTypeEnum{{}}),
	NamedArgs:       NO_NAMED_ARGS,
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

var FuncDebug = BuiltInFunc{
	Name:            FUNC_DEBUG,
	ReturnValues:    ZERO_RETURN_VALS,
	MinPosArgCount:  0,
	PosArgValidator: NewVarArgSchema([]RslTypeEnum{}),
	NamedArgs: map[string][]RslTypeEnum{
		namedArgEnd: {RslStringT},
		namedArgSep: {RslStringT},
	},
	Execute: func(f FuncInvocationArgs) []RslValue {
		RP.ScriptDebug(resolvePrintStr(f))
		return EMPTY
	},
}

func resolvePrintStr(f FuncInvocationArgs) string {
	var sb strings.Builder
	end := "\n"
	if endArg, ok := f.namedArgs[namedArgEnd]; ok {
		end = endArg.value.RequireStr(f.i, endArg.valueNode).String()
	}

	sep := " "
	if sepArg, ok := f.namedArgs[namedArgSep]; ok {
		sep = sepArg.value.RequireStr(f.i, sepArg.valueNode).String()
	}

	if len(f.args) == 0 {
		sb.WriteString(end)
	} else {
		for idx, v := range f.args {
			if v.value.Type() == RslStringT {
				// explicit handling for string so we don't print surrounding quotes when it's standalone
				sb.WriteString(ToPrintableQuoteStr(v.value.Val, false))
			} else {
				sb.WriteString(ToPrintableQuoteStr(v.value.Val, true))
			}
			if idx < len(f.args)-1 {
				sb.WriteString(sep)
			}
		}
		sb.WriteString(end)
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
