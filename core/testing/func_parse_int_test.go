package testing

import "testing"

func Test_ParseInt_Basic(t *testing.T) {
	rsl := `
a = parse_int("2")
print(a + 1)
a = parse_int("6178461748674861")
print(a + 1)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3\n6178461748674862\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_ParseInt_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = parse_int("asd")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/13 on 'parse_int': parse_int() could not parse \"asd\" as an integer\n")
	resetTestState()
}

func Test_ParseInt_ErrorsOnFloat(t *testing.T) {
	rsl := `
a = parse_int("2.4")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/13 on 'parse_int': parse_int() could not parse \"2.4\" as an integer\n")
	resetTestState()
}
