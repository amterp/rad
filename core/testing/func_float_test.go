package testing

import "testing"

func Test_Float_Basic(t *testing.T) {
	rsl := `
a = float("2.4")
print(a + 1.5)
a = float("123124.1232")
print(a + 1.5)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3.9\n123125.6232\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Float_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = float("asd")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/9 on 'float': float() could not parse \"asd\" as an float\n")
	resetTestState()
}

func Test_Float_CanParseInt(t *testing.T) {
	rsl := `
a = float("2")
print(a + 1.1)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3.1\n")
	assertNoErrors(t)
	resetTestState()
}
