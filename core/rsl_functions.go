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
	//funcName := function.GetLexeme()
	//namedArgsMap := toMap(i, namedArgs)
	//
	//switch funcName {
	//case KEYS:
	//	if len(args) != 1 {
	//		i.error(function, fmt.Sprintf("%s() takes exactly one argument", KEYS))
	//	}
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	validateExpectedNamedArgsOld(i, function, NO_NAMED_ARGS, namedArgsMap)
	//	switch coerced := args[0].(type) {
	//	case RslMapOld:
	//		return coerced.KeysGeneric()
	//	default:
	//		i.error(function, fmt.Sprintf("%s() takes a map, got %s", KEYS, TypeAsString(args[0])))
	//	}
	//case VALUES:
	//	if len(args) != 1 {
	//		i.error(function, fmt.Sprintf("%s() takes exactly one argument", VALUES))
	//	}
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	validateExpectedNamedArgsOld(i, function, NO_NAMED_ARGS, namedArgsMap)
	//	switch coerced := args[0].(type) {
	//	case RslMapOld:
	//		return coerced.Values()
	//	default:
	//		i.error(function, fmt.Sprintf("%s() takes a map, got %s", VALUES, TypeAsString(args[0])))
	//	}
	//case PARSE_JSON:
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	validateExpectedNamedArgsOld(i, function, NO_NAMED_ARGS, namedArgsMap)
	//	return runParseJson(i, function, args)
	//case HTTP_GET:
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	return runHttpGet(i, function, args, namedArgsMap)
	//case HTTP_POST:
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	return runHttpPost(i, function, args, namedArgsMap)
	//case HTTP_PUT:
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	return runHttpPut(i, function, args, namedArgsMap)
	//case PARSE_INT:
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, TWO_ARGS)
	//	validateExpectedNamedArgsOld(i, function, NO_NAMED_ARGS, namedArgsMap)
	//	return runParseInt(i, function, numExpectedReturnValues, args)
	//case PARSE_FLOAT:
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, TWO_ARGS)
	//	validateExpectedNamedArgsOld(i, function, NO_NAMED_ARGS, namedArgsMap)
	//	return runParseFloat(i, function, numExpectedReturnValues, args)
	//case ABS:
	//	assertExpectedNumReturnValuesOld(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	validateExpectedNamedArgsOld(i, function, NO_NAMED_ARGS, namedArgsMap)
	//	return runAbs(i, function, args)
	//default:
	//	//color, ok := ColorFromString(funcName)
	//	//if ok {
	//	//	assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, ONE_ARG)
	//	//	validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
	//	//	return runColor(i, function, args, color)
	//	//} else {
	//	//	i.error(function, fmt.Sprintf("Unknown function: %v", funcName))
	//	//	panic(UNREACHABLE)
	//	//} // TODO DELETE
	//}
	panic(UNREACHABLE)
}

func runAbs(i *MainInterpreter, function Token, args []interface{}) interface{} {
	if len(args) != 1 {
		i.error(function, ABS+fmt.Sprintf("() takes exactly one argument, got %d", len(args)))
	}

	switch coerced := args[0].(type) {
	case int64:
		return AbsInt(coerced)
	case float64:
		return AbsFloat(coerced)
	default:
		i.error(function, ABS+fmt.Sprintf("() takes an integer or float, got %s", TypeAsString(args[0])))
		panic(UNREACHABLE)
	}
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
