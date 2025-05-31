package testing

import "testing"

func Test_UpperLower(t *testing.T) {
	script := `
a = "aLiCe"
print(upper(a))
print(lower(a))`
	setupAndRunCode(t, script, "--color=never")
	expected := `ALICE
alice
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
