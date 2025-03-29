package testing

import "testing"

func Test_Func_Round_IntsWithPrecision(t *testing.T) {
	rsl := `
a = 1
b = 2
print(round(a, b))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Round_FloatsWithPrecision(t *testing.T) {
	rsl := `
a = 2.234
b = 1
print(round(a, b))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2.2\n")
	assertNoErrors(t)
}

func Test_Func_Round_FloatsWithZeroPrecision(t *testing.T) {
	rsl := `
a = 2.234
b = 0
print(round(a, b))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Func_Round_IntsWithoutPrecision(t *testing.T) {
	rsl := `
a = 1
print(round(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Round_FloatsWithoutPrecision(t *testing.T) {
	rsl := `
a = 2.234
print(round(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Func_Round_ErrorsPrecisionLessThan0(t *testing.T) {
	rsl := `
a = 1
b = -1
print(round(a, b))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L4:16

  print(round(a, b))
                 ^ Precision must be non-negative, got -1
`
	assertError(t, 1, expected)
}

func Test_Func_Round_ErrorsPrecisionString(t *testing.T) {
	rsl := `
a = 1
b = "ab"
print(round(a, b))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L4:16

  print(round(a, b))
                 ^ Got "string" as the 2nd argument of round(), but must be: int
`
	assertError(t, 1, expected)
}

func Test_Func_Round_ErrorsWithString(t *testing.T) {
	rsl := `
a = "ab"
b = 1
print(round(a, b))
	`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L4:13

  print(round(a, b))
              ^
              Got "string" as the 1st argument of round(), but must be: float or int
`
	assertError(t, 1, expected)
}
