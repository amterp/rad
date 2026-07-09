package core

import (
	"time"

	"github.com/amterp/rad/rts/rl"
)

var FuncFormatEpoch = BuiltInFunc{
	Name: FUNC_FORMAT_EPOCH,
	Execute: func(f FuncInvocation) RadValue {
		epoch := f.GetInt("_epoch")
		format := f.GetStr("_format").Plain()
		tz := f.GetStr("tz").Plain()
		unit := f.GetStr("unit").Plain()

		if format == "" {
			return f.ReturnErrf(rl.ErrCannotFormat, "Cannot format epoch with an empty format string")
		}

		isNegative := epoch < 0
		absEpoch := epoch
		if isNegative {
			absEpoch = -absEpoch
		}

		second, nanoSecond, _, errVal := resolveEpochUnit(f, FUNC_FORMAT_EPOCH, absEpoch, unit)
		if errVal != nil {
			return *errVal
		}

		if isNegative {
			second = -second
			nanoSecond = -nanoSecond
		}

		location, errVal := resolveTimeLocation(f, tz)
		if errVal != nil {
			return *errVal
		}

		goTime := time.Unix(second, nanoSecond).In(location)
		return f.Return(goTime.Format(convertFormatToGoLayout(format)))
	},
}
