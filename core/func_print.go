package core

import (
	"bytes"
	"strings"

	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/jsoncolor"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncPrint = BuiltInFunc{
	Name: FUNC_PRINT,
	Execute: func(f FuncInvocation) RadValue {
		RP.Print(resolvePrintStr(f))
		return VOID_SENTINEL
	},
}

var FuncPPrint = BuiltInFunc{
	Name: FUNC_PPRINT,
	Execute: func(f FuncInvocation) RadValue {
		item := f.GetArg("_item")

		jsonStruct := RadToJsonType(item)
		output := prettify(f.i, f.callNode, jsonStruct)
		RP.Printf(output)
		return VOID_SENTINEL
	},
}

var FuncDebug = BuiltInFunc{
	Name: FUNC_DEBUG,
	Execute: func(f FuncInvocation) RadValue {
		RP.ScriptDebug(resolvePrintStr(f))
		return VOID_SENTINEL
	},
}

var FuncPrintErr = BuiltInFunc{
	Name: FUNC_PRINT_ERR,
	Execute: func(f FuncInvocation) RadValue {
		RP.ScriptStderrf(resolvePrintStr(f))
		return VOID_SENTINEL
	},
}

func resolvePrintStr(f FuncInvocation) string {
	var sb strings.Builder
	items := f.GetList("_items")
	end := f.GetStr("end").Plain()
	sep := f.GetStr("sep").Plain()

	if items.IsEmpty() {
		sb.WriteString(end)
	} else {
		for idx, v := range items.Values {
			switch v.Type() {
			case rl.RadStrT, rl.RadErrorT:
				// explicit handling for string so we don't print surrounding quotes when it's standalone
				sb.WriteString(ToPrintableQuoteStr(v.Val, false))
			default:
				sb.WriteString(ToPrintableQuoteStr(v.Val, true))
			}
			if idx < items.LenInt()-1 {
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
