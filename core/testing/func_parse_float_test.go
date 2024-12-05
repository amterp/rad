package testing

import "testing"

func Test_ParseFloat_Basic(t *testing.T) {
	rsl := `
a = parse_float("2.4")
print(a + 1.5)
a = parse_float("123124.1232")
print(a + 1.5)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3.9\n123125.6232\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseFloat_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = parse_float("asd")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/15 on 'parse_float': parse_float() could not parse \"asd\" as an float\n")
	resetTestState()
}

func Test_ParseFloat_CanParseInt(t *testing.T) {
	rsl := `
a = parse_float("2")
print(a + 1.1)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3.1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseFloat_CanReadErrorIfNone(t *testing.T) {
	rsl := `
a, err = parse_float("2.4")
print(a + 1.5)
print(err)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3.9\n{}\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseFloat_CanReadErrorIfExists(t *testing.T) {
	rsl := `
a, err = parse_float("asd")
print(a)
print(err.msg)
print(err.code)
print(err)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "0\nparse_float() could not parse \"asd\" as an float\nRAD20002\n{ code: RAD20002, msg: parse_float() could not parse \"asd\" as an float }\n")
	assertNoErrors(t)
	resetTestState()
}
