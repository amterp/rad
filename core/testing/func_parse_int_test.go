package testing

import "testing"

func Test_ParseInt_Basic(t *testing.T) {
	rsl := `
a = parse_int("2")
print(a + 1)
a = parse_int("6178461748674861")
print(a + 1)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n6178461748674862\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseInt_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = parse_int("asd")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:5

  a = parse_int("asd")
      ^^^^^^^^^^^^^^^^ parse_int() failed to parse "asd"
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_ParseInt_ErrorsOnFloat(t *testing.T) {
	rsl := `
a = parse_int("2.4")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:5

  a = parse_int("2.4")
      ^^^^^^^^^^^^^^^^ parse_int() failed to parse "2.4"
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_ParseInt_CanReadErrorIfNone(t *testing.T) {
	rsl := `
a, err = parse_int("2")
print(a + 1)
print(err)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n{ }\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseInt_CanReadErrorIfExists(t *testing.T) {
	rsl := `
a, err = parse_int("asd")
print(a)
print(err.msg)
print(err.code)
print(err)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `0
parse_int() failed to parse "asd"
RAD20001
{ "code": "RAD20001", "msg": "parse_int() failed to parse "asd"" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseInt_DoesNotErrorIfOutputNotRead(t *testing.T) {
	rsl := `
parse_int("2")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "")
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseInt_ErrorsIfExpectingTooManyReturnValues(t *testing.T) {
	rsl := `
a, b, c = parse_int("2.4")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:11

  a, b, c = parse_int("2.4")
            ^^^^^^^^^^^^^^^^ parse_int() returns 1 or 2 values, but 3 are expected
`
	assertError(t, 1, expected)
	resetTestState()
}
