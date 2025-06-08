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
	PosArgValidator: NewVarArgSchema([]RadTypeEnum{}),
	NamedArgs: map[string][]RadTypeEnum{
		namedArgEnd: {RadStringT},
		namedArgSep: {RadStringT},
	},
	Execute: func(f FuncInvocationArgs) []RadValue {
		RP.Printf(resolvePrintStr(f))
		return EMPTY
	},
}

var FuncPPrint = BuiltInFunc{
	Name:            FUNC_PPRINT,
	ReturnValues:    ZERO_RETURN_VALS,
	MinPosArgCount:  0,
	PosArgValidator: NewEnumerableArgSchema([][]RadTypeEnum{{}}),
	NamedArgs:       NO_NAMED_ARGS,
	Execute: func(f FuncInvocationArgs) []RadValue {
		if len(f.args) == 0 {
			RP.Printf("\n")
		}

		arg := f.args[0]
		jsonStruct := RadToJsonType(arg.value)
		output := prettify(f.i, f.callNode, jsonStruct)
		RP.Printf(output)
		return EMPTY
	},
}

var FuncDebug = BuiltInFunc{
	Name:            FUNC_DEBUG,
	ReturnValues:    ZERO_RETURN_VALS,
	MinPosArgCount:  0,
	PosArgValidator: NewVarArgSchema([]RadTypeEnum{}),
	NamedArgs: map[string][]RadTypeEnum{
		namedArgEnd: {RadStringT},
		namedArgSep: {RadStringT},
	},
	Execute: func(f FuncInvocationArgs) []RadValue {
		RP.ScriptDebug(resolvePrintStr(f))
		return EMPTY
	},
}

var FuncPrintErr = BuiltInFunc{
	Name:            FUNC_PRINT_ERR,
	ReturnValues:    ZERO_RETURN_VALS,
	MinPosArgCount:  0,
	PosArgValidator: NewVarArgSchema([]RadTypeEnum{}),
	NamedArgs: map[string][]RadTypeEnum{
		namedArgEnd: {RadStringT},
		namedArgSep: {RadStringT},
	},
	Execute: func(f FuncInvocationArgs) []RadValue {
		RP.ScriptStderrf(resolvePrintStr(f))
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
			switch v.value.Type() {
			case RadStringT, RadErrorT:
				// explicit handling for string so we don't print surrounding quotes when it's standalone
				sb.WriteString(ToPrintableQuoteStr(v.value.Val, false))
			default:
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
