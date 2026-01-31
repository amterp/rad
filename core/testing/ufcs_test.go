package testing

import "testing"

func Test_Ufcs_Basic(t *testing.T) {
	script := `
print("hi".upper())
`
	setupAndRunCode(t, script, "--color=never")
	expected := `HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_Chained(t *testing.T) {
	script := `
"hi".upper().print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_WithArgs(t *testing.T) {
	script := `
"hello!".replace("l", "o").print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `heooo!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_ChainedMultiline(t *testing.T) {
	t.Skip("TODO this is not supported yet, but should be")
	script := `
"hi"
	.upper()
	.print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_ErrorsIfIncorrectUfcsArg(t *testing.T) {
	script := `
a = [1, 2, 3]
a[1].replace("l", "o")
`
	setupAndRunCode(t, script, "--color=never")
	// todo ^ known issue, we pass a bad node for the error pointing
	assertErrorContains(t, 1, "RAD30001", "Value '2' (int) is not compatible with expected type 'str'")
}
