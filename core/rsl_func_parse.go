package core

import (
	"fmt"
	"strconv"
)

func runParseInt(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, PARSE_INT+fmt.Sprintf("() takes 1 argument, got %d", len(args)))
	}

	switch coerced := args[0].(type) {
	case RslString:
		str := coerced.Plain()
		parsed, err := strconv.Atoi(str)
		if err != nil {
			i.error(function, PARSE_INT+fmt.Sprintf("() could not parse %q as an integer", str))
		}
		return int64(parsed)
	default:
		i.error(function, PARSE_INT+fmt.Sprintf("() takes a string, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}
