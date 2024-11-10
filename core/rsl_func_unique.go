package core

import (
	"fmt"
	"github.com/samber/lo"
)

// todo allow string input? to avoid needing to split(input, "") first
func runUnique(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, UNIQUE+fmt.Sprintf("() takes 1 argument, got %d", len(args)))
	}

	switch arr := args[0].(type) {
	case []interface{}:
		return lo.Uniq(arr)
	default:
		i.error(function, UNIQUE+fmt.Sprintf("() takes an array as its argument, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}
