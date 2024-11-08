package testing

import "testing"

func TestDate(t *testing.T) {
	rsl := `
a = now_date()
print(a)
print(a + "100")
`
	setupAndRunCode(t, rsl)
	expected := `2019-12-13
2019-12-13100
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestYear(t *testing.T) {
	rsl := `
a = now_year()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `2019
2119
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMonth(t *testing.T) {
	rsl := `
a = now_month()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `12
112
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestDay(t *testing.T) {
	rsl := `
a = now_day()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `13
113
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestHour(t *testing.T) {
	rsl := `
a = now_hour()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `14
114
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMinute(t *testing.T) {
	rsl := `
a = now_minute()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `15
115
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestSecond(t *testing.T) {
	rsl := `
a = now_second()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `16
116
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestEpochSeconds(t *testing.T) {
	rsl := `
a = epoch_seconds()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `1576246516
1576246616
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestEpochMillis(t *testing.T) {
	rsl := `
a = epoch_millis()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `1576246516123
1576246516223
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestEpochNanos(t *testing.T) {
	rsl := `
a = epoch_nanos()
print(a)
print(a + 100)
`
	setupAndRunCode(t, rsl)
	expected := `1576246516123123123
1576246516123123223
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
