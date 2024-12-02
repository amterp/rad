package core

import (
	"fmt"
	"github.com/samber/lo"
	"sort"
	"strings"
)

var (
	NO_NAMED_ARGS = make([]string, 0)
)

// RunRslNonVoidFunction returns pointers to values e.g. *string
func RunRslNonVoidFunction(
	i *MainInterpreter,
	function Token,
	numExpectedReturnValues int,
	args []interface{},
	namedArgs []NamedArg,
) interface{} {
	funcName := function.GetLexeme()
	namedArgsMap := toMap(i, namedArgs)

	switch funcName {
	case "len":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runLen(i, function, args)
	case "now_date": // todo is this name good? current_date? date?
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return RClock.Now().Format("2006-01-02")
	case "now_year":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return int64(RClock.Now().Year())
	case "now_month":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return int64(RClock.Now().Month())
	case "now_day":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return int64(RClock.Now().Day())
	case "now_hour":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return int64(RClock.Now().Hour())
	case "now_minute":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return int64(RClock.Now().Minute())
	case "now_second":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return int64(RClock.Now().Second())
	case "epoch_seconds":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return RClock.Now().Unix()
	case "epoch_millis":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return RClock.Now().UnixMilli()
	case "epoch_nanos":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return RClock.Now().UnixNano()
	case "replace":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runReplace(i, function, args)
	case "join":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return RunJoin(i, function, args)
	case "upper":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		arg := args[0]
		switch coerced := arg.(type) {
		case RslString:
			return coerced.Upper()
		default:
			return strings.ToUpper(ToPrintable(arg))
		}
	case "lower":
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		arg := args[0]
		switch coerced := arg.(type) {
		case RslString:
			return coerced.Lower()
		default:
			return strings.ToLower(ToPrintable(arg))
		}
	case "starts_with":
		if len(args) != 2 {
			i.error(function, "starts_with() takes exactly two arguments")
		}
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return strings.HasPrefix(ToPrintable(args[0]), ToPrintable(args[1]))
	case "ends_with":
		if len(args) != 2 {
			i.error(function, "ends_with() takes exactly two arguments")
		}
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return strings.HasSuffix(ToPrintable(args[0]), ToPrintable(args[1]))
	case PICK:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		return runPick(i, function, args, namedArgsMap)
	case PICK_KV:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		return runPickKv(i, function, args, namedArgsMap)
	case PICK_FROM_RESOURCE:
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap) // todo add 'prompt'
		return runPickFromResource(i, function, args, numExpectedReturnValues)
	case KEYS:
		if len(args) != 1 {
			i.error(function, fmt.Sprintf("%s() takes exactly one argument", KEYS))
		}
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		switch coerced := args[0].(type) {
		case RslMap:
			return coerced.KeysGeneric()
		default:
			i.error(function, fmt.Sprintf("%s() takes a map, got %s", KEYS, TypeAsString(args[0])))
		}
	case VALUES:
		if len(args) != 1 {
			i.error(function, fmt.Sprintf("%s() takes exactly one argument", VALUES))
		}
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		switch coerced := args[0].(type) {
		case RslMap:
			return coerced.Values()
		default:
			i.error(function, fmt.Sprintf("%s() takes a map, got %s", VALUES, TypeAsString(args[0])))
		}
	case RAND:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runRand(i, function, args)
	case RAND_INT:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runRandInt(i, function, args)
	case TRUNCATE:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runTruncate(i, function, args)
	case SPLIT:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runSplit(i, function, args)
	case RANGE:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runRange(i, function, args)
	case UNIQUE:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runUnique(i, function, args)
	case SORT_FUNC:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		return runSort(i, function, args, namedArgsMap)
	case CONFIRM:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runConfirm(i, function, args)
	case PARSE_JSON:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
		return runParseJson(i, function, args)
	case HTTP_GET:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		return runHttpGet(i, function, args, namedArgsMap)
	case HTTP_POST:
		assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
		return runHttpPost(i, function, args, namedArgsMap)
	default:
		color, ok := ColorFromString(funcName)
		if ok {
			assertExpectedNumReturnValues(i, function, funcName, numExpectedReturnValues, 1)
			validateExpectedNamedArgs(i, function, NO_NAMED_ARGS, namedArgsMap)
			return runColor(i, function, args, color)
		} else {
			i.error(function, fmt.Sprintf("Unknown function: %v", funcName))
			panic(UNREACHABLE)
		}
	}
	panic(UNREACHABLE)
}

func RunRslFunction(i *MainInterpreter, call FunctionCall) {
	funcToken := call.Function
	args := evalArgs(i, call.Args)
	functionName := funcToken.GetLexeme()
	namedArgsMap := toMap(i, call.NamedArgs)

	switch functionName {
	case PRINT:
		validateExpectedNamedArgs(i, call.Function, NO_NAMED_ARGS, namedArgsMap) // todo implement coloring?
		runPrint(args)
	case PPRINT:
		if len(args) > 1 {
			i.error(funcToken, PPRINT+"() takes zero or one argument")
		}
		validateExpectedNamedArgs(i, call.Function, NO_NAMED_ARGS, namedArgsMap)
		runPrettyPrint(i, funcToken, args)
	case DEBUG:
		validateExpectedNamedArgs(i, call.Function, NO_NAMED_ARGS, namedArgsMap)
		runDebug(args)
	case EXIT:
		// todo allow following exit code with msg?
		validateExpectedNamedArgs(i, call.Function, NO_NAMED_ARGS, namedArgsMap)
		runExit(i, funcToken, args)
	case SLEEP:
		runSleep(i, funcToken, args, namedArgsMap)
	case SEED_RANDOM:
		validateExpectedNamedArgs(i, call.Function, NO_NAMED_ARGS, namedArgsMap)
		runSeedRandom(i, funcToken, args)
	default:
		RunRslNonVoidFunction(i, funcToken, NO_NUM_RETURN_VALUES_CONSTRAINT, args, call.NamedArgs)
	}
}

func runLen(i *MainInterpreter, function Token, values []interface{}) int64 {
	if len(values) != 1 {
		i.error(function, "len() takes exactly one argument")
	}
	switch v := values[0].(type) {
	case RslString:
		return v.Len()
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

	subject := ToPrintable(values[0]) // todo should assert only strings on subject
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

func validateExpectedNamedArgs(i *MainInterpreter, function Token, expectedArgs []string, namedArgs map[string]interface{}) {
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
