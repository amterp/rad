package testing

import "testing"

func Test_Int_Basic(t *testing.T) {
	rsl := `
a = int("2")
print(a + 1)
a = int("6178461748674861")
print(a + 1)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3\n6178461748674862\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Int_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = int("asd")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/7 on 'int': int() could not parse \"asd\" as an integer\n")
	resetTestState()
}

func Test_Int_ErrorsOnFloat(t *testing.T) {
	rsl := `
a = int("2.4")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/7 on 'int': int() could not parse \"2.4\" as an integer\n")
	resetTestState()
}
