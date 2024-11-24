package core

import (
	"fmt"
)

// todo allow string input? to avoid needing to split(input, "") first
func runUnique(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, UNIQUE+fmt.Sprintf("() takes 1 argument, got %d", len(args)))
	}

	switch arr := args[0].(type) {
	case []interface{}:
		return uniq(arr)
	default:
		i.error(function, UNIQUE+fmt.Sprintf("() takes an array as its argument, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}

func uniq(arr []interface{}) interface{} {
	seen := make(map[string]struct{})
	result := make([]interface{}, 0)

	for _, item := range arr {
		key := ToPrintable(item) // a little eh...
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}

	return result
}
