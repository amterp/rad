package testing

import "testing"

func Test_Func_Pow_Integers(t *testing.T) {
	script := `
print(pow(2, 3))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "8\n")
	assertNoErrors(t)
}

func Test_Func_Pow_FloatBase(t *testing.T) {
	script := `
print(pow(2.5, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "6.25\n")
	assertNoErrors(t)
}

func Test_Func_Pow_FloatExponent(t *testing.T) {
	script := `
print(pow(4, 0.5))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Func_Pow_NegativeBase(t *testing.T) {
	script := `
print(pow(-2, 3))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-8\n")
	assertNoErrors(t)
}

func Test_Func_Pow_NegativeExponent(t *testing.T) {
	script := `
print(pow(2, -2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0.25\n")
	assertNoErrors(t)
}

func Test_Func_Pow_ZeroExponent(t *testing.T) {
	script := `
print(pow(5, 0))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Pow_ZeroBase(t *testing.T) {
	script := `
print(pow(0, 3))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0\n")
	assertNoErrors(t)
}

func Test_Func_Pow_OneBase(t *testing.T) {
	script := `
print(pow(1, 100))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Pow_FractionalExponent(t *testing.T) {
	script := `
print(pow(8, 0.333333))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1.9999986137061192\n")
	assertNoErrors(t)
}

func Test_Func_Pow_ErrorsWithStringBase(t *testing.T) {
	script := `
print(pow("abc", 2))
	`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:11

  print(pow("abc", 2))
            ^^^^^ Value '"abc"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}

func Test_Func_Pow_ErrorsWithStringExponent(t *testing.T) {
	script := `
print(pow(2, "abc"))
	`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:14

  print(pow(2, "abc"))
               ^^^^^
               Value '"abc"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}