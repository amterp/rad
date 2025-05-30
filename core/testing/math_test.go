package testing

import "testing"

func Test_Math_Float(t *testing.T) {
	script := `
print(1.2 + 2.3)
print(3.0 / 2.0)
print(3.0 / 2)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `3.5
1.5
1.5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Math_Int(t *testing.T) {
	script := `
print(1 + 3)
print(3 / 2)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `4
1.5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Math_ErrorsOnIntIntDivisionByZero(t *testing.T) {
	script := `
a = 1 / 0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:9

  a = 1 / 0
          ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_ErrorsOnFloatIntDivisionByZero(t *testing.T) {
	script := `
a = 1.0 / 0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:11

  a = 1.0 / 0
            ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_ErrorsOnIntFloatDivisionByZero(t *testing.T) {
	script := `
a = 1 / 0.0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:9

  a = 1 / 0.0
          ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_ErrorsOnFloatFloatDivisionByZero(t *testing.T) {
	script := `
a = 1.0 / 0.0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:11

  a = 1.0 / 0.0
            ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_CompoundDivideByZeroIntIntErrors(t *testing.T) {
	script := `
a = 1
a /= 0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:6

  a /= 0
       ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_CompoundDivideByZeroIntFloatErrors(t *testing.T) {
	script := `
a = 1
a /= 0.0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:6

  a /= 0.0
       ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_CompoundDivideByZeroFloatIntErrors(t *testing.T) {
	script := `
a = 1.0
a /= 0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:6

  a /= 0
       ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_CompoundDivideByZeroFloatFloatErrors(t *testing.T) {
	script := `
a = 1.0
a /= 0.0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:6

  a /= 0.0
       ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
}

func Test_Math_CanHaveManyPluses(t *testing.T) {
	script := `
a = 1 +++++++++ +2
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Math_CanHaveManyMinuses(t *testing.T) {
	script := `
a = 1 + -------2
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1\n")
	assertNoErrors(t)
}
