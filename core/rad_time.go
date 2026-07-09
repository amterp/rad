package core

import (
	"fmt"
	"strconv"
	"time"

	"github.com/amterp/rad/rts/rl"
)

func NewTimeMap(time time.Time) *RadMap {
	timeMap := NewRadMap()
	hour := time.Hour()
	minute := time.Minute()
	second := time.Second()

	// ISO 8601 weekday: Monday=1 .. Sunday=7 (Go's Weekday has Sunday=0)
	weekday := int(time.Weekday())
	if weekday == 0 {
		weekday = 7
	}

	timeMap.SetPrimitiveStr("date", time.Format("2006-01-02"))
	timeMap.SetPrimitiveInt("year", time.Year())
	timeMap.SetPrimitiveInt("month", int(time.Month()))
	timeMap.SetPrimitiveInt("day", time.Day())
	timeMap.SetPrimitiveInt("weekday", weekday)
	timeMap.SetPrimitiveInt("hour", hour)
	timeMap.SetPrimitiveInt("minute", minute)
	timeMap.SetPrimitiveInt("second", second)
	timeMap.SetPrimitiveStr("time", fmt.Sprintf("%02d:%02d:%02d", hour, minute, second))

	epochMap := NewRadMap()
	epochMap.SetPrimitiveInt64("seconds", time.Unix())
	epochMap.SetPrimitiveInt64("millis", time.UnixMilli())
	epochMap.SetPrimitiveInt64("nanos", time.UnixNano())

	timeMap.SetPrimitiveMap("epoch", epochMap)

	return timeMap
}

// resolveEpochUnit interprets absEpoch (non-negative) in the given unit ("auto"
// detects by digit count) as seconds + nanos. fracMultiplier converts a
// fractional part of that unit to nanos, for callers accepting float epochs.
// A non-nil errVal is ready to return from the calling function as-is.
func resolveEpochUnit(f FuncInvocation, funcName string, absEpoch int64, unit string) (second int64, nanoSecond int64, fracMultiplier float64, errVal *RadValue) {
	if unit == constAuto {
		digitCount := len(strconv.FormatInt(absEpoch, 10))
		switch digitCount {
		case 1, 2, 3, 4, 5, 6, 7, 8, 9, 10:
			second = absEpoch
			nanoSecond = 0
			fracMultiplier = 1e9
		case 13:
			second = absEpoch / 1_000
			nanoSecond = (absEpoch % 1_000) * 1_000_000
			fracMultiplier = 1e6
		case 16:
			second = absEpoch / 1_000_000
			nanoSecond = (absEpoch % 1_000_000) * 1_000
			fracMultiplier = 1e3
		case 19:
			second = absEpoch / 1_000_000_000
			nanoSecond = absEpoch % 1_000_000_000
			fracMultiplier = 1
		default:
			errMsg := fmt.Sprintf(
				"Ambiguous epoch length (%d digits). Use '%s' to disambiguate.",
				digitCount,
				namedArgUnit,
			)
			err := f.Return(NewErrorStr(errMsg).SetCode(rl.ErrAmbiguousEpoch).SetSpan(nodeSpanPtr(f.callNode)))
			return 0, 0, 0, &err
		}
	} else {
		switch unit {
		case constSeconds:
			second = absEpoch
			nanoSecond = 0
			fracMultiplier = 1e9
		case constMillis:
			second = absEpoch / 1_000
			nanoSecond = (absEpoch % 1_000) * 1_000_000
			fracMultiplier = 1e6
		case constMicros:
			second = absEpoch / 1_000_000
			nanoSecond = (absEpoch % 1_000_000) * 1_000
			fracMultiplier = 1e3
		case constNanos:
			second = absEpoch / 1_000_000_000
			nanoSecond = absEpoch % 1_000_000_000
			fracMultiplier = 1
		case constMilliseconds, constMicroseconds, constNanoseconds:
			replacements := map[string]string{
				constMilliseconds: constMillis,
				constMicroseconds: constMicros,
				constNanoseconds:  constNanos,
			}
			f.i.emitErrorWithHint(rl.ErrInvalidTimeUnit, f.callNode,
				fmt.Sprintf("%s unit %q is no longer valid", funcName, unit),
				fmt.Sprintf("Unit names were shortened in v0.9. Use %q instead. See: https://amterp.dev/rad/migrations/v0.9/", replacements[unit]))
			panic(UNREACHABLE)
		default:
			err := f.ReturnErrf(rl.ErrInvalidTimeUnit,
				"invalid units %q; expected one of %s, %s, %s, %s, %s",
				unit, constAuto, constSeconds, constMillis, constMicros, constNanos)
			return 0, 0, 0, &err
		}
	}
	return second, nanoSecond, fracMultiplier, nil
}

// resolveTimeLocation resolves a user-provided tz string ("local" or an IANA
// name) to a location. A non-nil errVal is ready to return as-is.
func resolveTimeLocation(f FuncInvocation, tz string) (*time.Location, *RadValue) {
	if tz == "local" {
		return RClock.Local(), nil
	}
	location, err := time.LoadLocation(tz)
	if err != nil {
		errVal := f.ReturnErrf(rl.ErrInvalidTimeZone, "invalid time zone %q", tz)
		return nil, &errVal
	}
	return location, nil
}
