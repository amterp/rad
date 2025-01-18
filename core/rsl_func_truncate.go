package core

import "fmt"

func runTruncate(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 2 {
		i.error(function, TRUNCATE+fmt.Sprintf("() takes 2 arguments, got %d", len(args)))
	}

	// todo if you have a list of strings, you lose the attributes (should be returning a RslString?) RAD-109
	str := ToPrintable(args[0])
	switch coerced := args[1].(type) {
	case int64:
		if coerced < 0 {
			i.error(function, TRUNCATE+fmt.Sprintf("() takes a non-negative int, got %d", coerced))
		}
		if coerced >= int64(StrLen(str)) {
			return args[0]
		}
		if terminalSupportsUtf8 {
			str = str[:coerced-1]
			str += "…"
		} else {
			str = str[:coerced-3]
			str += "..."
		}
		return NewRslString(str)
	default:
		i.error(function, TRUNCATE+fmt.Sprintf("() takes an int as its second arg, got %s", TypeAsString(args[1])))
		panic(UNREACHABLE)
	}
}
