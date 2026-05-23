package core

import (
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
			err := sleep(f.i, time.Duration(coerced)*time.Second, f.namedArgs)
			if err != nil {
				return f.Return(err)
			}
			return VOID_SENTINEL
		case float64:
			err := sleep(f.i, time.Duration(coerced*1000)*time.Millisecond, f.namedArgs)
			if err != nil {
				return f.Return(err)
			}
			return VOID_SENTINEL
		case RadString:
			durStr := coerced.Plain()

			// Bare numeric string -> interpret as seconds
			floatVal, err := rts.ParseFloat(durStr)
			if err == nil {
				err := sleep(f.i, time.Duration(floatVal*1000)*time.Millisecond, f.namedArgs)
				if err != nil {
					return f.Return(err)
				}
				return VOID_SENTINEL
			}

			// Human-readable duration string
			nanos, err := ParseDurationString(durStr)
			if err == nil {
				err := sleep(f.i, time.Duration(nanos), f.namedArgs)
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

func sleep(i *Interpreter, dur time.Duration, namedArgs map[string]namedArg) *RadError {
	if dur < 0 {
		return NewErrorStrf("Cannot take a negative duration: %q", dur.String())
	}

	if title, ok := namedArgs[namedArgTitle]; ok {
		RP.Printf(ToPrintableQuoteStr(title.value, false) + "\n")
	}

	// Use the interpreter's signal context so a signal handler can wake us
	// up early. If the sleep returns due to ctx cancellation, the signal
	// will be dispatched at the next checkpoint after this builtin returns.
	RSleep(i.signals.Ctx(), dur)
	return nil
}
