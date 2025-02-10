package core

import (
	"fmt"
	"sort"
	"strings"

	"github.com/samber/lo"
)

// RunRslNonVoidFunction returns pointers to values e.g. *string
func RunRslNonVoidFunction(
	i *MainInterpreter,
	function Token,
	numExpectedReturnValues int,
	args []interface{},
	namedArgs []NamedArg,
) interface{} {
	output := runRslNonVoidFunction(i, function, numExpectedReturnValues, args, namedArgs)
	bugCheckOutputType(i, function, output)
	return output
}

func runRslNonVoidFunction(i *MainInterpreter, function Token, numExpectedReturnValues int, args []interface{}, namedArgs []NamedArg) interface{} {
	panic(UNREACHABLE)
}

func evalArgs(i *MainInterpreter, args []Expr) []interface{} {
	var values []interface{}
	for _, v := range args {
		val := v.Accept(i)
		values = append(values, val)
	}
	return values
}

func validateExpectedNamedArgsOld(i *MainInterpreter, function Token, expectedArgs []string, namedArgs map[string]interface{}) {
	var unknownArgs []string
	for k := range namedArgs {
		if !lo.Contains(expectedArgs, k) {
			unknownArgs = append(unknownArgs, k)
		}
	}

	if len(unknownArgs) == 0 {
		return
	}

	sort.Strings(unknownArgs)
	unknownArgsStr := strings.Join(unknownArgs, ", ")
	i.error(function, fmt.Sprintf("Unknown named argument(s): %s", unknownArgsStr))
}

func bugCheckOutputType(i *MainInterpreter, token Token, output interface{}) {
	switch output.(type) {
	case RslString, int64, float64, bool, []interface{}, RslMapOld:
		return
	default:
		i.error(token, fmt.Sprintf("Bug! Unexpected return type: %v", TypeAsString(output)))
	}
}
