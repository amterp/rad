package testing

import "testing"

func Test_Ufcs_Basic(t *testing.T) {
	rsl := `
print("hi".upper())
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_Chained(t *testing.T) {
	rsl := `
"hi".upper().print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_WithArgs(t *testing.T) {
	rsl := `
"hello!".replace("l", "o").print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `heooo!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_ChainedMultiline(t *testing.T) {
	t.Skip("TODO this is not supported yet, but should be")
	rsl := `
"hi"
	.upper()
	.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `HI
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Ufcs_ErrorsIfIncorrectUfcsArg(t *testing.T) {
	rsl := `
a = [1, 2, 3]
a[1].replace("l", "o")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:1

  a[1].replace("l", "o")
  ^ Got "int" as the 1st argument of replace(), but must be: string
`
	// todo ^ known issue, we pass a bad node for the error pointing
	assertError(t, 1, expected)
}
