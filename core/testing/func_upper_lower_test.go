package testing

import "testing"

func Test_UpperLower(t *testing.T) {
	rsl := `
a = "aLiCe"
print(upper(a))
print(lower(a))
print(upper(5))
print(lower(5))`
	setupAndRunCode(t, rsl)
	expected := `ALICE
alice
5
5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
