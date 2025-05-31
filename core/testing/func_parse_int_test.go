package testing

import "testing"

func Test_ParseInt_Basic(t *testing.T) {
	script := `
a = parse_int("2")
print(a + 1)
a = parse_int("6178461748674861")
print(a + 1)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n6178461748674862\n")
	assertNoErrors(t)
}

func Test_ParseInt_ErrorsOnAlphabetical(t *testing.T) {
	script := `
a = parse_int("asd")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  a = parse_int("asd")
      ^^^^^^^^^^^^^^^^ parse_int() failed to parse "asd"
`
	assertError(t, 1, expected)
}

func Test_ParseInt_ErrorsOnFloat(t *testing.T) {
	script := `
a = parse_int("2.4")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  a = parse_int("2.4")
      ^^^^^^^^^^^^^^^^ parse_int() failed to parse "2.4"
`
	assertError(t, 1, expected)
}

func Test_ParseInt_CanReadErrorIfNone(t *testing.T) {
	script := `
a, err = parse_int("2")
print(a + 1)
print(err)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n{ }\n")
	assertNoErrors(t)
}

func Test_ParseInt_CanReadErrorIfExists(t *testing.T) {
	script := `
a, err = parse_int("asd")
print(a)
print(err.msg)
print(err.code)
print(err)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0
parse_int() failed to parse "asd"
RAD20001
{ "code": "RAD20001", "msg": "parse_int() failed to parse "asd"" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ParseInt_DoesNotErrorIfOutputNotRead(t *testing.T) {
	script := `
parse_int("2")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "")
	assertNoErrors(t)
}

func Test_ParseInt_ErrorsIfExpectingTooManyReturnValues(t *testing.T) {
	script := `
a, b, c = parse_int("2.4")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  a, b, c = parse_int("2.4")
  ^^^^^^^^^^^^^^^^^^^^^^^^^^ Cannot assign 2 values to 3 variables
`
	assertError(t, 1, expected)
}
