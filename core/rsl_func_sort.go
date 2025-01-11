package core

import (
	"fmt"

	"github.com/samber/lo"
)

const (
	SORT_REVERSE = "reverse"
)

func runSort(i *MainInterpreter, function Token, args []interface{}, namedArgs map[string]interface{}) []interface{} {
	if len(args) != 1 {
		i.error(function, SORT_FUNC+fmt.Sprintf("() takes exactly 1 positional arg, got %d", len(args)))
	}

	validateExpectedNamedArgs(i, function, []string{SORT_REVERSE}, namedArgs)
	parsedArgs := parseSortArgs(i, function, namedArgs)

	switch coerced := args[0].(type) {
	case []interface{}:
		return sortList(i, function, coerced, lo.Ternary(parsedArgs.Reverse, Desc, Asc))
	default:
		i.error(function, SORT_FUNC+fmt.Sprintf("() takes a list, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
}

func parseSortArgs(i *MainInterpreter, function Token, args map[string]interface{}) SortNamedArgs {
	parsedArgs := SortNamedArgs{
		Reverse: false,
	}

	if reverse, ok := args[SORT_REVERSE]; ok {
		if parsedArgs.Reverse, ok = reverse.(bool); !ok {
			i.error(function, SORT_FUNC+fmt.Sprintf("() %s must be a boolean, got %s", SORT_REVERSE, TypeAsString(reverse)))
		}
	}

	return parsedArgs
}

type SortNamedArgs struct {
	Reverse bool
}
