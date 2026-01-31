package testing

import "testing"

func Test_FuncParam_ScientificNotation_Int_Valid_1e6(t *testing.T) {
	script := `
fn test(x: int = 1e6):
    print("{x}")

test()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1000000\n")
	assertNoErrors(t)
}

func Test_FuncParam_ScientificNotation_Int_Valid_1_2e10(t *testing.T) {
	script := `
fn test(x: int = 1.2e10):
    print("{x}")

test()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "12000000000\n")
	assertNoErrors(t)
}

func Test_FuncParam_ScientificNotation_Int_Valid_1000e_minus_2(t *testing.T) {
	script := `
fn test(x: int = 1000e-2):
    print("{x}")

test()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10\n")
	assertNoErrors(t)
}

func Test_FuncParam_ScientificNotation_Int_Invalid_1e_minus_5(t *testing.T) {
	script := `
fn test(x: int = 1e-5):
    print("{x}")

test()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1,
		"fn test(x: int = 1e-5):",
		"Scientific notation value does not evaluate to a whole number",
	)
}

func Test_FuncParam_ScientificNotation_Int_Invalid_1_25e1(t *testing.T) {
	script := `
fn test(x: int = 1.25e1):
    print("{x}")

test()
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1,
		"fn test(x: int = 1.25e1):",
		"Scientific notation value does not evaluate to a whole number",
	)
}

func Test_FuncParam_ScientificNotation_Float_Valid_1e6(t *testing.T) {
	script := `
fn test(x: float = 1e6):
    print("{x}")

test()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1000000\n")
	assertNoErrors(t)
}

func Test_FuncParam_ScientificNotation_Float_Valid_1e_minus_5(t *testing.T) {
	script := `
fn test(x: float = 1e-5):
    print("{x}")

test()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0.00001\n")
	assertNoErrors(t)
}

func Test_FuncParam_ScientificNotation_Lambda_Valid(t *testing.T) {
	script := `
f = fn(x: int = 1e6) x * 2
print("{f()}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2000000\n")
	assertNoErrors(t)
}

func Test_FuncParam_ScientificNotation_Lambda_Invalid(t *testing.T) {
	script := `
f = fn(x: int = 1e-5) x * 2
print("{f()}")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1,
		"f = fn(x: int = 1e-5) x * 2",
		"Scientific notation value does not evaluate to a whole number",
	)
}
