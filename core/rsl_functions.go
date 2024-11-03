package core

import (
	"fmt"
	"strings"
)

// RunRslNonVoidFunction returns pointers to values e.g. *string
func RunRslNonVoidFunction(
	i *MainInterpreter,
	function Token,
	numExpectedReturnValues int,
	args []interface{},
	namedArgs []NamedArg,
) interface{} {
	functionName := function.GetLexeme()

	switch functionName {
	case "len":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return runLen(i, function, args)
	case "today_date": // todo is this name good? current_date? date?
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return RClock.Now().Format("2006-01-02")
	case "today_year":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return int64(RClock.Now().Year())
	case "today_month":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return int64(RClock.Now().Month())
	case "today_day":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return int64(RClock.Now().Day())
	case "today_hour":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return int64(RClock.Now().Hour())
	case "today_minute":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return int64(RClock.Now().Minute())
	case "today_second":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return int64(RClock.Now().Second())
	case "epoch_seconds":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return RClock.Now().Unix()
	case "epoch_millis":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return RClock.Now().UnixMilli()
	case "epoch_nanos":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return RClock.Now().UnixNano()
	case "replace":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return runReplace(i, function, args)
	case "join":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return RunJoin(i, function, args)
	case "upper":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return strings.ToUpper(ToPrintable(args[0]))
	case "lower":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return strings.ToLower(ToPrintable(args[0]))
	case "starts_with":
		if len(args) != 2 {
			i.error(function, "starts_with() takes exactly two arguments")
		}
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return strings.HasPrefix(ToPrintable(args[0]), ToPrintable(args[1]))
	case "ends_with":
		if len(args) != 2 {
			i.error(function, "ends_with() takes exactly two arguments")
		}
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return strings.HasSuffix(ToPrintable(args[0]), ToPrintable(args[1]))
	case "contains":
		if len(args) != 2 {
			i.error(function, "contains() takes exactly two arguments")
		}
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return strings.Contains(ToPrintable(args[0]), ToPrintable(args[1]))
	case "pick":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return runPick(i, function, args)
	case PICK_KV:
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return runPickKv(i, function, args)
	case PICK_FROM_RESOURCE:
		return runPickFromResource(i, function, args, numExpectedReturnValues)
	case KEYS:
		if len(args) != 1 {
			i.error(function, fmt.Sprintf("%s() takes exactly one argument", KEYS))
		}
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		switch coerced := args[0].(type) {
		case RslMap:
			return coerced.KeysGeneric()
		default:
			i.error(function, fmt.Sprintf("%s() takes a map, got %T", KEYS, args[0]))
		}
	case VALUES:
		if len(args) != 1 {
			i.error(function, fmt.Sprintf("%s() takes exactly one argument", VALUES))
		}
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		switch coerced := args[0].(type) {
		case RslMap:
			return coerced.Values()
		default:
			i.error(function, fmt.Sprintf("%s() takes a map, got %T", VALUES, args[0]))
		}
	default:
		i.error(function, fmt.Sprintf("Unknown function: %v", functionName))
		panic(UNREACHABLE)
	}
	panic(UNREACHABLE)
}

func RunRslFunction(i *MainInterpreter, call FunctionCall) {
	funcToken := call.Function
	args := evalArgs(i, call.Args)
	functionName := funcToken.GetLexeme()

	switch functionName {
	case PRINT:
		runPrint(args)
	case PPRINT:
		if len(args) > 1 {
			i.error(funcToken, PPRINT+"() takes zero or one argument")
		}
		runPrettyPrint(i, funcToken, args)
	case DEBUG:
		runDebug(args)
	case EXIT:
		runExit(i, funcToken, args)
	case SLEEP:
		runSleep(i, funcToken, args, toMap(i, call.NamedArgs))
	default:
		RunRslNonVoidFunction(i, funcToken, NO_NUM_RETURN_VALUES_CONSTRAINT, args, call.NamedArgs)
	}
}

func runLen(i *MainInterpreter, function Token, values []interface{}) int64 {
	if len(values) != 1 {
		i.error(function, "len() takes exactly one argument")
	}
	switch v := values[0].(type) {
	case string:
		return int64(len(v))
	case []interface{}:
		return int64(len(v))
	case RslMap:
		return int64(v.Len())
	default:
		i.error(function, "len() takes a string or collection")
		panic(UNREACHABLE)
	}
}

func runReplace(i *MainInterpreter, function Token, values []interface{}) interface{} {
	if len(values) != 3 {
		i.error(function, "replace() takes exactly three arguments")
	}

	subject := ToPrintable(values[0])
	oldRegex := ToPrintable(values[1])
	newRegex := ToPrintable(values[2])

	return Replace(i, function, subject, oldRegex, newRegex)
}

func evalArgs(i *MainInterpreter, args []Expr) []interface{} {
	var values []interface{}
	for _, v := range args {
		val := v.Accept(i)
		values = append(values, val)
	}
	return values
}

func toMap(i *MainInterpreter, args []NamedArg) map[string]interface{} {
	m := make(map[string]interface{})
	for _, arg := range args {
		m[arg.Arg.GetLexeme()] = arg.Value.Accept(i)
	}
	return m
}

func assertExpectedNumReturnValues(
	i *MainInterpreter,
	function Token,
	funcName string,
	numExpectedReturnValues int,
	actualNumReturnValues int,
) {
	if numExpectedReturnValues != NO_NUM_RETURN_VALUES_CONSTRAINT && numExpectedReturnValues != actualNumReturnValues {
		i.error(function, fmt.Sprintf("%v() returns %v return values, but %v are expected",
			funcName, actualNumReturnValues, numExpectedReturnValues))
	}
}
