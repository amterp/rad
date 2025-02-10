package testing

import "testing"

func Test_Func_Now(t *testing.T) {
	rsl := `
a = now()
print(a)

print(a.date, type_of(a.date))
print(a.year, type_of(a.year))
print(a.month, type_of(a.month))
print(a.day, type_of(a.day))
print(a.hour, type_of(a.hour))
print(a.minute, type_of(a.minute))
print(a.second, type_of(a.second))

print(a.epoch.seconds, type_of(a.epoch.seconds))
print(a.epoch.millis, type_of(a.epoch.millis))
print(a.epoch.nanos, type_of(a.epoch.nanos))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `{ "date": "2019-12-13", "year": 2019, "month": 12, "day": 13, "hour": 14, "minute": 15, "second": 16, "epoch": { "seconds": 1576246516, "millis": 1576246516123, "nanos": 1576246516123123123 } }
2019-12-13 string
2019 int
12 int
13 int
14 int
15 int
16 int
1576246516 int
1576246516123 int
1576246516123123123 int
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
