package core

import (
	"strings"
	"time"

	"github.com/amterp/rad/rts"

	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncSleep = BuiltInFunc{
	Name: FUNC_SLEEP,
	Execute: func(f FuncInvocationArgs) RadValue {
		arg := f.args[0]
		switch coerced := arg.value.Val.(type) {
		case int64:
			sleep(f.i, arg.node, time.Duration(coerced)*time.Second, f.namedArgs)
			return VOID_SENTINEL
		case float64:
			sleep(f.i, arg.node, time.Duration(coerced*1000)*time.Millisecond, f.namedArgs)
			return VOID_SENTINEL
		case RadString:
			durStr := strings.Replace(coerced.Plain(), " ", "", -1)

			floatVal, err := rts.ParseFloat(durStr)
			if err == nil {
				sleep(f.i, arg.node, time.Duration(floatVal*1000)*time.Millisecond, f.namedArgs)
				return VOID_SENTINEL
			}

			dur, err := time.ParseDuration(durStr)
			if err == nil {
				sleep(f.i, arg.node, dur, f.namedArgs)
				return VOID_SENTINEL
			}

			f.i.errorf(arg.node, "Invalid string argument: %q", coerced.Plain())
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
		RP.Printf(ToPrintableQuoteStr(title.value, false) + "\n")
	}

	RSleep(dur)
}
