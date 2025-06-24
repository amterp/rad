package testing

import "testing"

func Test_Func_Ceil_Ints(t *testing.T) {
	script := `
print(ceil(1))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Ceil_Negative_Ints(t *testing.T) {
	script := `
print(ceil(-1))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1\n")
	assertNoErrors(t)
}

func Test_Func_Ceil_Floats(t *testing.T) {
	script := `
print(ceil(2.234))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Func_Ceil_Negative_Floats(t *testing.T) {
	script := `
print(ceil(-2.234))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
}

func Test_Func_Ceil_Errors_With_String(t *testing.T) {
	script := `
print(ceil("ab"))
	`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:12

  print(ceil("ab"))
             ^^^^ Value '"ab"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}
