package testing

import "testing"

func TestConvertDuration_SecondsToMinutes(t *testing.T) {
	setupAndRunCode(t, `print(convert_duration(90, "seconds").minutes)`)
	assertOnlyOutput(t, stdOutBuffer, "1.5\n")
	assertNoErrors(t)
}

func TestConvertDuration_MillisToSeconds(t *testing.T) {
	setupAndRunCode(t, `print(convert_duration(1500, "millis").seconds)`)
	assertOnlyOutput(t, stdOutBuffer, "1.5\n")
	assertNoErrors(t)
}

func TestConvertDuration_DaysToHours(t *testing.T) {
	setupAndRunCode(t, `print(convert_duration(1, "days").hours)`)
	assertOnlyOutput(t, stdOutBuffer, "24\n")
	assertNoErrors(t)
}

func TestConvertDuration_FractionalHoursToMinutes(t *testing.T) {
	setupAndRunCode(t, `print(convert_duration(1.5, "hours").minutes)`)
	assertOnlyOutput(t, stdOutBuffer, "90\n")
	assertNoErrors(t)
}

func TestConvertDuration_Zero(t *testing.T) {
	script := `
d = convert_duration(0, "seconds")
print(d.nanos)
print(d.millis)
print(d.hours)
`
	setupAndRunCode(t, script)
	assertOnlyOutput(t, stdOutBuffer, "0\n0\n0\n")
	assertNoErrors(t)
}

func TestConvertDuration_Negative(t *testing.T) {
	setupAndRunCode(t, `print(convert_duration(-60, "seconds").minutes)`)
	assertOnlyOutput(t, stdOutBuffer, "-1\n")
	assertNoErrors(t)
}

func TestConvertDuration_NanosToMicros(t *testing.T) {
	setupAndRunCode(t, `print(convert_duration(5000, "nanos").micros)`)
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
}

func TestConvertDuration_MicrosToMillis(t *testing.T) {
	setupAndRunCode(t, `print(convert_duration(2500, "micros").millis)`)
	assertOnlyOutput(t, stdOutBuffer, "2.5\n")
	assertNoErrors(t)
}
