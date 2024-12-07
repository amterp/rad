package core

import (
	"fmt"
)

// todo behaves different from python's, need to justify those or correct it
func runParseJson(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, PARSE_JSON+fmt.Sprintf("() takes exactly one argument, got %d", len(args)))
	}

	switch coerced := args[0].(type) {
	case RslString:
		out, err := TryConvertJsonToNativeTypes(i, function, coerced.Plain())
		if err != nil {
			i.error(function, fmt.Sprintf("Error parsing JSON: %v", err))
		}
		return out
	default:
		// maybe a bit harsh, should allow just passthrough of e.g. int64?
		i.error(function, PARSE_JSON+fmt.Sprintf("() expects string, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}
