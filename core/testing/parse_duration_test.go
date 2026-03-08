package testing

import "testing"

func TestParseDuration_BasicSeconds(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("5s").seconds)`)
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
}

func TestParseDuration_MinutesAndSeconds(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("5m23s").millis)`)
	assertOnlyOutput(t, stdOutBuffer, "323000\n")
	assertNoErrors(t)
}

func TestParseDuration_Milliseconds(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("300ms").millis)`)
	assertOnlyOutput(t, stdOutBuffer, "300\n")
	assertNoErrors(t)
}

func TestParseDuration_FractionalHours(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("1.5h").minutes)`)
	assertOnlyOutput(t, stdOutBuffer, "90\n")
	assertNoErrors(t)
}

func TestParseDuration_Days(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("3d5h").hours)`)
	assertOnlyOutput(t, stdOutBuffer, "77\n")
	assertNoErrors(t)
}

func TestParseDuration_OneDayOnly(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("1d").hours)`)
	assertOnlyOutput(t, stdOutBuffer, "24\n")
	assertNoErrors(t)
}

func TestParseDuration_FractionalDays(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("0.5d").hours)`)
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
}

func TestParseDuration_NanosPrecision(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("5m30ns").nanos)`)
	assertOnlyOutput(t, stdOutBuffer, "300000000030\n")
	assertNoErrors(t)
}

func TestParseDuration_Negative(t *testing.T) {
	script := `
d = parse_duration("-5m")
print(d.minutes)
print(d.nanos)
`
	setupAndRunCode(t, script)
	assertOnlyOutput(t, stdOutBuffer, "-5\n-300000000000\n")
	assertNoErrors(t)
}

func TestParseDuration_Spaces(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("5m 23s").millis)`)
	assertOnlyOutput(t, stdOutBuffer, "323000\n")
	assertNoErrors(t)
}

func TestParseDuration_Error(t *testing.T) {
	setupAndRunCode(t, `parse_duration("invalid!")`, "--color=never")
	expected := `error[RAD20043]: Failed to parse duration "invalid!": invalid duration: time: invalid duration "invalid!"
  --> TestCase:1:1
  |
1 | parse_duration("invalid!")
  | ^^^^^^^^^^^^^^^^^^^^^^^^^^
  |
   = info: rad explain RAD20043

`
	assertError(t, 1, expected)
}

func TestParseDuration_DayHourMinute(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("1d12h30m").millis)`)
	assertOnlyOutput(t, stdOutBuffer, "131400000\n")
	assertNoErrors(t)
}

func TestParseDuration_NanosField(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration("1s").nanos)`)
	assertOnlyOutput(t, stdOutBuffer, "1000000000\n")
	assertNoErrors(t)
}

func TestParseDuration_ErrorBareMinus(t *testing.T) {
	setupAndRunCode(t, `parse_duration("-")`, "--color=never")
	assertErrorContains(t, 1, "RAD20043", "empty duration string")
}

func TestParseDuration_ErrorDuplicateDays(t *testing.T) {
	setupAndRunCode(t, `parse_duration("1d2d")`, "--color=never")
	assertErrorContains(t, 1, "RAD20043", "multiple day components")
}

func TestParseDuration_ErrorOverflow(t *testing.T) {
	setupAndRunCode(t, `parse_duration("106751d24h")`, "--color=never")
	assertErrorContains(t, 1, "RAD20043", "overflow")
}

func TestParseDuration_DotFiveDays(t *testing.T) {
	setupAndRunCode(t, `print(parse_duration(".5d").hours)`)
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
}
