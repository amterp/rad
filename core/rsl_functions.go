package core

import (
	"fmt"
	"os"
	"strings"
)

// RunRslNonVoidFunction returns pointers to values e.g. *string
func RunRslNonVoidFunction(
	i *MainInterpreter,
	function Token,
	numExpectedReturnValues int,
	args []interface{},
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
	case "pick_kv":
		assertExpectedNumReturnValues(i, function, functionName, numExpectedReturnValues, 1)
		return runPickKv(i, function, args)
	case "pick_from_resource":
		return runPickFromResource(i, function, args, numExpectedReturnValues)
	default:
		i.error(function, fmt.Sprintf("Unknown function: %v", functionName))
		panic(UNREACHABLE)
	}
}

func RunRslFunction(i *MainInterpreter, function Token, args []interface{}) {
	functionName := function.GetLexeme()
	switch functionName {
	case "print": // todo would be nice to make this a reference to a var that GoLand can find
		runPrint(args)
	case "debug":
		runDebug(args)
	case "exit":
		if len(args) == 0 {
			os.Exit(0)
		} else if len(args) == 1 {
			arg := args[0]
			switch coerced := arg.(type) {
			case int64:
				os.Exit(int(coerced))
			default:
				i.error(function, "exit() takes an integer argument")
			}
		} else {
			i.error(function, "exit() takes zero or one argument")
		}
	default:
		RunRslNonVoidFunction(i, function, NO_NUM_RETURN_VALUES_CONSTRAINT, args)
	}
}

func runPrint(values []interface{}) {
	output := resolveOutputString(values)
	RP.Print(output)
}

func runDebug(values []interface{}) {
	output := resolveOutputString(values)
	RP.ScriptDebug(output)
}

func resolveOutputString(values []interface{}) string {
	output := ""

	if len(values) == 0 {
		output = "\n"
	} else {
		for _, v := range values {
			output += ToPrintable(v) + " "
		}
		output = output[:len(output)-1] // remove last space
		output = output + "\n"
	}
	return output
}

func runLen(i *MainInterpreter, function Token, values []interface{}) int64 {
	if len(values) != 1 {
		i.error(function, "len() takes exactly one argument")
	}
	switch v := values[0].(type) {
	case string:
		return int64(len(v))
	case []string:
		return int64(len(v))
	case []int64:
		return int64(len(v))
	case []float64:
		return int64(len(v))
	case []interface{}:
		return int64(len(v))
	default:
		i.error(function, "len() takes a string or array")
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
