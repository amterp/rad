package core

import (
	"strconv"
	"strings"
	"time"

	ts "github.com/tree-sitter/go-tree-sitter"
)

const (
	namedArgTitle = "title"
)

var FuncSleep = Func{
	Name:             FUNC_SLEEP,
	ReturnValues:     ZERO_RETURN_VALS,
	RequiredArgCount: 1,
	ArgTypes:         [][]RslTypeEnum{{RslIntT, RslFloatT, RslStringT}},
	NamedArgs: map[string][]RslTypeEnum{
		namedArgTitle: {RslStringT},
	},
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, namedArgs map[string]namedArg) []RslValue {
		arg := args[0]
		switch coerced := arg.value.Val.(type) {
		case int64:
			sleep(i, arg.node, time.Duration(coerced)*time.Second, namedArgs)
			return EMPTY
		case float64:
			sleep(i, arg.node, time.Duration(coerced*1000)*time.Millisecond, namedArgs)
			return EMPTY
		case RslString:
			durStr := strings.Replace(coerced.Plain(), " ", "", -1)

			floatVal, err := strconv.ParseFloat(durStr, 64)
			if err == nil {
				sleep(i, arg.node, time.Duration(floatVal*1000)*time.Millisecond, namedArgs)
				return EMPTY
			}

			dur, err := time.ParseDuration(durStr)
			if err == nil {
				sleep(i, arg.node, dur, namedArgs)
				return EMPTY
			}

			i.errorf(arg.node, "Invalid string argument: %q", coerced.Plain())
			panic(UNREACHABLE)
		default:
			bugIncorrectTypes(FUNC_SLEEP)
			panic(UNREACHABLE)
		}
	},
}

func sleep(i *Interpreter, argNode *ts.Node, dur time.Duration, namedArgs map[string]namedArg) {
	if dur < 0 {
		i.errorf(argNode, "%s() cannot take a negative duration: %q", FUNC_SLEEP, dur.String())
	}

	if title, ok := namedArgs[namedArgTitle]; ok {
		RP.Print(ToPrintableQuoteStr(title.value, false) + "\n")
	}

	RSleep(dur)
}
