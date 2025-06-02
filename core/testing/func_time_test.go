package testing

import "testing"

func Test_Func_Now(t *testing.T) {
	script := `
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
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "date": "2019-12-13", "year": 2019, "month": 12, "day": 13, "hour": 14, "minute": 15, "second": 16, "time": "14:15:16", "epoch": { "seconds": 1576206916, "millis": 1576206916123, "nanos": 1576206916123123123 } }
2019-12-13 string
2019 int
12 int
13 int
14 int
15 int
16 int
1576206916 int
1576206916123 int
1576206916123123123 int
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_NowWithTimeZone(t *testing.T) {
	script := `
a, b = now(tz="America/Chicago")
print(b)
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
	setupAndRunCode(t, script, "--color=never")
	expected := `null
{ "date": "2019-12-12", "year": 2019, "month": 12, "day": 12, "hour": 21, "minute": 15, "second": 16, "time": "21:15:16", "epoch": { "seconds": 1576206916, "millis": 1576206916123, "nanos": 1576206916123123123 } }
2019-12-12 string
2019 int
12 int
12 int
21 int
15 int
16 int
1576206916 int
1576206916123 int
1576206916123123123 int
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_NowErrorsForInvalidTimeZone(t *testing.T) {
	script := `
a, b = now(tz="invalid time zone")
print(a, b)
a = now(tz="another bad one")
`

	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `null { "code": "RAD20009", "msg": "Invalid time zone 'invalid time zone'" }
`
	expectedStderr := `Error at L4:5

  a = now(tz="another bad one")
      ^^^^^^^^^^^^^^^^^^^^^^^^^ Invalid time zone 'another bad one'
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertExitCode(t, 1)
}

func Test_Func_ParseEpochSeconds(t *testing.T) {
	script := `
a, b = parse_epoch(1712345678)
print(b)
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
{ "date": "2024-04-06", "year": 2024, "month": 4, "day": 6, "hour": 6, "minute": 34, "second": 38, "time": "06:34:38", "epoch": { "seconds": 1712345678, "millis": 1712345678000, "nanos": 1712345678000000000 } }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ParseEpochMillis(t *testing.T) {
	script := `
a, b = parse_epoch(1712345678123)
print(b)
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
{ "date": "2024-04-06", "year": 2024, "month": 4, "day": 6, "hour": 6, "minute": 34, "second": 38, "time": "06:34:38", "epoch": { "seconds": 1712345678, "millis": 1712345678123, "nanos": 1712345678123000000 } }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ParseEpochMicros(t *testing.T) {
	script := `
a, b = parse_epoch(1712345678123123)
print(b)
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
{ "date": "2024-04-06", "year": 2024, "month": 4, "day": 6, "hour": 6, "minute": 34, "second": 38, "time": "06:34:38", "epoch": { "seconds": 1712345678, "millis": 1712345678123, "nanos": 1712345678123123000 } }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ParseEpochNanos(t *testing.T) {
	script := `
a, b = parse_epoch(1712345678123123123)
print(b)
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
{ "date": "2024-04-06", "year": 2024, "month": 4, "day": 6, "hour": 6, "minute": 34, "second": 38, "time": "06:34:38", "epoch": { "seconds": 1712345678, "millis": 1712345678123, "nanos": 1712345678123123123 } }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ParseEpochErrorsIfAmbiguous(t *testing.T) {
	script := `
a, b = parse_epoch(17123456781)
print(b)
print(a)
a = parse_epoch(17123456781)
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `{ "code": "RAD20007", "msg": "Ambiguous epoch length (11 digits). Use 'unit' to disambiguate." }
null
`
	expectedStderr := `Error at L5:17

  a = parse_epoch(17123456781)
                  ^^^^^^^^^^^
                  Ambiguous epoch length (11 digits). Use 'unit' to disambiguate.
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertExitCode(t, 1)
}

func Test_Func_ParseEpochTimeZoneNegativeAmbiguousButWithUnits(t *testing.T) {
	script := `
a, b = parse_epoch(-17123456781, unit = "milliseconds")
print(b)
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
{ "date": "1969-06-17", "year": 1969, "month": 6, "day": 17, "hour": 5, "minute": 29, "second": 3, "time": "05:29:03", "epoch": { "seconds": -17123457, "millis": -17123456781, "nanos": -17123456781000000 } }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ParseEpochTimeZone(t *testing.T) {
	script := `
a, b = parse_epoch(1712345678123, tz = "America/Chicago")
print(b)
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
{ "date": "2024-04-05", "year": 2024, "month": 4, "day": 5, "hour": 14, "minute": 34, "second": 38, "time": "14:34:38", "epoch": { "seconds": 1712345678, "millis": 1712345678123, "nanos": 1712345678123000000 } }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
