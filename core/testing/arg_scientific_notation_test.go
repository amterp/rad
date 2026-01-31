package testing

import "testing"

func Test_Arg_ScientificNotation_Int_Valid_1e6(t *testing.T) {
	script := `
args:
	num int = 1e6
print("{num}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1000000\n")
	assertNoErrors(t)
}

func Test_Arg_ScientificNotation_Int_Valid_1000e_minus_2(t *testing.T) {
	script := `
args:
	num int = 1000e-2
print("{num}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10\n")
	assertNoErrors(t)
}

func Test_Arg_ScientificNotation_Int_Valid_1_2e10(t *testing.T) {
	script := `
args:
	num int = 1.2e10
print("{num}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "12000000000\n")
	assertNoErrors(t)
}

func Test_Arg_ScientificNotation_Int_Invalid_1e_minus_5(t *testing.T) {
	script := `
args:
	num int = 1e-5
print("{num}")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "Scientific notation value does not evaluate to a whole number")
}

func Test_Arg_ScientificNotation_Int_Invalid_1_25e1(t *testing.T) {
	script := `
args:
	num int = 1.25e1
print("{num}")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "Scientific notation value does not evaluate to a whole number")
}

func Test_Arg_ScientificNotation_Float_Valid_1e6(t *testing.T) {
	script := `
args:
	num float = 1e6
print("{num}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1000000\n")
	assertNoErrors(t)
}

func Test_Arg_ScientificNotation_Float_Valid_1e_minus_5(t *testing.T) {
	script := `
args:
	num float = 1e-5
print("{num}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0.00001\n")
	assertNoErrors(t)
}
