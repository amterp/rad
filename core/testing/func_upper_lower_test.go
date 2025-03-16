package testing

import "testing"

func Test_UpperLower(t *testing.T) {
	rsl := `
a = "aLiCe"
print(upper(a))
print(lower(a))`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `ALICE
alice
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
