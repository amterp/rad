package testing

import "testing"

func Test_ParseFloat_Basic(t *testing.T) {
	script := `
a = parse_float("2.4")
print(a + 1.5)
a = parse_float("123124.1232")
print(a + 1.5)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3.9\n123125.6232\n")
	assertNoErrors(t)
}

func Test_ParseFloat_ErrorsOnAlphabetical(t *testing.T) {
	script := `
a = parse_float("asd")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20002", "parse_float() failed to parse")
}

func Test_ParseFloat_CanParseInt(t *testing.T) {
	script := `
a = parse_float("2")
print(a + 1.1)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3.1\n")
	assertNoErrors(t)
}

func Test_ParseFloat_CanCatchEvenIfNoError(t *testing.T) {
	script := `
a = parse_float("2.4") catch:
	pass
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2.4\n")
	assertNoErrors(t)
}

func Test_ParseFloat_CaCatchErrorIfExists(t *testing.T) {
	script := `
a = parse_float("asd") catch:
	print("Got", a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Got parse_float() failed to parse "asd"
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
