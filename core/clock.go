package core

import "time"

type Clock interface {
	Now() time.Time
	Local() *time.Location
}

type RealClock struct {
}

func NewRealClock() Clock {
	return &RealClock{}
}

func (r *RealClock) Now() time.Time {
	return time.Now()
}

func (r *RealClock) Local() *time.Location {
	return time.Local
}

type FixedClock struct {
	NowTime time.Time
}

func NewFixedClock(year, month, day, hour, minute, second, nano int64, tz *time.Location) Clock {
	return &FixedClock{
		NowTime: time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), int(nano), tz),
	}
}

func (f *FixedClock) Now() time.Time {
	return f.NowTime
}

func (f *FixedClock) Local() *time.Location {
	return f.NowTime.Location()
}
