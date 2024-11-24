package core

import (
	"encoding/json"
	"fmt"
)

func runParseJson(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, PARSE_JSON+fmt.Sprintf("() takes exactly one argument, got %d", len(args)))
	}

	switch coerced := args[0].(type) {
	case RslString:
		var m interface{}
		err := json.Unmarshal([]byte(coerced.Plain()), &m)
		if err != nil {
			i.error(function, fmt.Sprintf("Error parsing JSON: %v", err))
		}
		return ConvertToNativeTypes(i, function, m)
	default:
		// maybe a bit harsh, should allow just passthrough of e.g. int64?
		i.error(function, PARSE_JSON+fmt.Sprintf("() expects string, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}
