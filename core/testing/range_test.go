package testing

import "testing"

func TestRange_Int_OnlyEnd(t *testing.T) {
	rsl := `
a = range(10)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 ]\n")
	assertNoErrors(t)
}

func TestRange_Int_StartEnd(t *testing.T) {
	rsl := `
a = range(0, 10)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 ]\n")
	assertNoErrors(t)
}

func TestRange_Int_NonZeroStart(t *testing.T) {
	rsl := `
a = range(20, 30)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 20, 21, 22, 23, 24, 25, 26, 27, 28, 29 ]\n")
	assertNoErrors(t)
}

func TestRange_Int_NegativeRange(t *testing.T) {
	rsl := `
a = range(-20, -10)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ -20, -19, -18, -17, -16, -15, -14, -13, -12, -11 ]\n")
	assertNoErrors(t)
}

func TestRange_Int_Steps(t *testing.T) {
	rsl := `
a = range(20, 40, 2)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 20, 22, 24, 26, 28, 30, 32, 34, 36, 38 ]\n")
	assertNoErrors(t)
}

func TestRange_Int_NegativeSteps(t *testing.T) {
	rsl := `
a = range(40, 20, -2)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 40, 38, 36, 34, 32, 30, 28, 26, 24, 22 ]\n")
	assertNoErrors(t)
}

func TestRange_Float_OnlyEndExclusive(t *testing.T) {
	rsl := `
a = range(10.0)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 ]\n")
	assertNoErrors(t)
}

func TestRange_Float_OnlyEndAbove(t *testing.T) {
	rsl := `
a = range(10.5)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10 ]\n")
	assertNoErrors(t)
}

func TestRange_Float_StartEnd(t *testing.T) {
	rsl := `
a = range(0.0, 10.0)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 0, 1, 2, 3, 4, 5, 6, 7, 8, 9 ]\n")
	assertNoErrors(t)
}

func TestRange_Float_NonZeroStart(t *testing.T) {
	rsl := `
a = range(20.5, 30.5)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 20.5, 21.5, 22.5, 23.5, 24.5, 25.5, 26.5, 27.5, 28.5, 29.5 ]\n")
	assertNoErrors(t)
}

func TestRange_Float_NegativeRange(t *testing.T) {
	rsl := `
a = range(-20.0, -10.0)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ -20, -19, -18, -17, -16, -15, -14, -13, -12, -11 ]\n")
	assertNoErrors(t)
}

func TestRange_Float_Steps(t *testing.T) {
	rsl := `
a = range(20, 40, 2.5)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 20, 22.5, 25, 27.5, 30, 32.5, 35, 37.5 ]\n")
	assertNoErrors(t)
}

func TestRange_Float_NegativeSteps(t *testing.T) {
	rsl := `
a = range(40, 20, -2.5)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 40, 37.5, 35, 32.5, 30, 27.5, 25, 22.5 ]\n")
	assertNoErrors(t)
}

func TestRange_CanGenerateLargeArray(t *testing.T) {
	rsl := `
a = range(100000)
print(len(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "100000\n")
	assertNoErrors(t)
}
