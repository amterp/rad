package testing

import "testing"

func Test_Func_Ceil_Ints(t *testing.T) {
	rsl := `
print(ceil(1))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Ceil_Negative_Ints(t *testing.T) {
	rsl := `
print(ceil(-1))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Ceil_Floats(t *testing.T) {
	rsl := `
print(ceil(2.234))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Ceil_Negative_Floats(t *testing.T) {
	rsl := `
print(ceil(-2.234))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Ceil_Errors_With_String(t *testing.T) {
	rsl := `
print(ceil("ab"))
	`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:12

  print(ceil("ab"))
             ^^^^
             Got "string" as the 1st argument of ceil(), but must be: float or int
`
	assertError(t, 1, expected)
	resetTestState()
}
