package core

import (
	"math"
	"time"

	"github.com/amterp/rad/rts/rl"
)

var FuncConvertDuration = BuiltInFunc{
	Name: FUNC_CONVERT_DURATION,
	Execute: func(f FuncInvocation) RadValue {
		value := f.GetFloat("_value")
		unit := f.GetStr("_unit").Plain()

		var nanosFloat float64
		switch unit {
		case "nanos":
			nanosFloat = value
		case "micros":
			nanosFloat = value * float64(time.Microsecond)
		case "millis":
			nanosFloat = value * float64(time.Millisecond)
		case "seconds":
			nanosFloat = value * float64(time.Second)
		case "minutes":
			nanosFloat = value * float64(time.Minute)
		case "hours":
			nanosFloat = value * float64(time.Hour)
		case "days":
			nanosFloat = value * 24 * float64(time.Hour)
		default:
			bugIncorrectTypes(FUNC_CONVERT_DURATION)
			panic(UNREACHABLE)
		}

		if math.IsNaN(nanosFloat) || math.IsInf(nanosFloat, 0) || math.Abs(nanosFloat) > float64(math.MaxInt64) {
			return f.ReturnErrf(rl.ErrNumInvalidRange, "Duration overflow: value too large for %s", unit)
		}

		return f.Return(NewDurationMap(int64(nanosFloat)))
	},
}
