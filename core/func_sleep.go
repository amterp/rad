package core

import (
	"strings"
	"time"

	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/rad/rts"
)

var FuncSleep = BuiltInFunc{
	Name: FUNC_SLEEP,
	Execute: func(f FuncInvocation) RadValue {
		duration := f.GetArg("_duration")
		switch coerced := duration.Val.(type) {
		case int64:
			err := sleep(time.Duration(coerced)*time.Second, f.namedArgs)
			if err != nil {
				return f.Return(err)
			}
			return VOID_SENTINEL
		case float64:
			err := sleep(time.Duration(coerced*1000)*time.Millisecond, f.namedArgs)
			if err != nil {
				return f.Return(err)
			}
			return VOID_SENTINEL
		case RadString:
			durStr := strings.Replace(coerced.Plain(), " ", "", -1)

			floatVal, err := rts.ParseFloat(durStr)
			if err == nil {
				err := sleep(time.Duration(floatVal*1000)*time.Millisecond, f.namedArgs)
				if err != nil {
					return f.Return(err)
				}
				return VOID_SENTINEL
			}

			dur, err := time.ParseDuration(durStr)
			if err == nil {
				err := sleep(dur, f.namedArgs)
				if err != nil {
					return f.Return(err)
				}
				return VOID_SENTINEL
			}

			return f.ReturnErrf(rl.ErrSleepStr, "Invalid string argument: %q", coerced.Plain())
		default:
			bugIncorrectTypes(FUNC_SLEEP)
			panic(UNREACHABLE)
		}
	},
}

func sleep(dur time.Duration, namedArgs map[string]namedArg) *RadError {
	if dur < 0 {
		return NewErrorStrf("Cannot take a negative duration: %q", dur.String())
	}

	if title, ok := namedArgs[namedArgTitle]; ok {
		RP.Printf(ToPrintableQuoteStr(title.value, false) + "\n")
	}

	RSleep(dur)
	return nil
}
