package core

import "fmt"

func runTruncate(i *MainInterpreter, function Token, args []interface{}) string {
	if len(args) != 2 {
		i.error(function, TRUNCATE+fmt.Sprintf("() takes 2 arguments, got %d", len(args)))
	}

	str := ToPrintable(args[0])
	switch coerced := args[1].(type) {
	case int64:
		if coerced < 0 {
			i.error(function, TRUNCATE+fmt.Sprintf("() takes a non-negative int, got %d", coerced))
		}
		if coerced >= int64(StrLen(str)) {
			return str
		}
		if isTerminalUtf8 {
			str = str[:coerced-1]
			str += "…"
		} else {
			str = str[:coerced-3]
			str += "..."
		}
		return str
	default:
		i.error(function, TRUNCATE+fmt.Sprintf("() takes an int as its second arg, got %s", TypeAsString(args[1])))
		panic(UNREACHABLE)
	}
}
