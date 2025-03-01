package testing

import "testing"

func Test_UpperLower(t *testing.T) {
	rsl := `
a = "aLiCe"
print(upper(a))
print(lower(a))`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `ALICE
alice
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
