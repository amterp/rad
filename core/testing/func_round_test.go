package testing

import "testing"

func Test_Func_Round_IntsWithPrecision(t *testing.T) {
	script := `
a = 1
b = 2
print(round(a, b))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Round_FloatsWithPrecision(t *testing.T) {
	script := `
a = 2.234
b = 1
print(round(a, b))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2.2\n")
	assertNoErrors(t)
}

func Test_Func_Round_FloatsWithZeroPrecision(t *testing.T) {
	script := `
a = 2.234
b = 0
c = round(a, b)
print(c, type_of(c))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2 int\n")
	assertNoErrors(t)
}

func Test_Func_Round_IntsWithoutPrecision(t *testing.T) {
	script := `
a = 1
print(round(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Round_FloatsWithoutPrecision(t *testing.T) {
	script := `
a = 2.234
b = round(a)
print(b, type_of(b))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2 int\n")
	assertNoErrors(t)
}

func Test_Func_Round_ErrorsPrecisionLessThan0(t *testing.T) {
	script := `
a = 1
b = -1
print(round(a, b))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L4:7

  print(round(a, b))
        ^^^^^^^^^^^ Precision must be non-negative, got -1 (RAD20017)
`
	assertError(t, 1, expected)
}

func Test_Func_Round_ErrorsPrecisionString(t *testing.T) {
	script := `
a = 1
b = "ab"
print(round(a, b))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L4:16

  print(round(a, b))
                 ^ Value '"ab"' (str) is not compatible with expected type 'int'
`
	assertError(t, 1, expected)
}

func Test_Func_Round_ErrorsWithString(t *testing.T) {
	script := `
a = "ab"
b = 1
print(round(a, b))
	`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L4:13

  print(round(a, b))
              ^ Value '"ab"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}
