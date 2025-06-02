package core

import (
	"fmt"
	"time"
)

func NewTimeMap(time time.Time) *RadMap {
	timeMap := NewRadMap()
	hour := time.Hour()
	minute := time.Minute()
	second := time.Second()

	timeMap.SetPrimitiveStr("date", time.Format("2006-01-02"))
	timeMap.SetPrimitiveInt("year", time.Year())
	timeMap.SetPrimitiveInt("month", int(time.Month()))
	timeMap.SetPrimitiveInt("day", time.Day())
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
