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
      ^^^^^^^^^^^^^^^^ parse_int() failed to parse "asd" (RAD20001)
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
      ^^^^^^^^^^^^^^^^ parse_int() failed to parse "2.4" (RAD20001)
`
	assertError(t, 1, expected)
}

func Test_ParseInt_CanCatchEvenIfNoError(t *testing.T) {
	script := `
a = parse_int("2") catch:
	print("Should not be here")
print("Got: {a}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Got: 2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ParseInt_CanReadErrorIfExists(t *testing.T) {
	script := `
a = parse_int("asd") catch:
	pass
print("Got: {a}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Got: parse_int() failed to parse "asd"
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
