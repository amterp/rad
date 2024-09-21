package testing

import "testing"

func TestDate(t *testing.T) {
	rsl := `
a = today_date()
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
a = today_year()
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
a = today_month()
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
a = today_day()
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
