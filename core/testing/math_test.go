package testing

import "testing"

func Test_Math_Float(t *testing.T) {
	rsl := `
print(1.2 + 2.3)
print(3.0 / 2.0)
print(3.0 / 2)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `3.5
1.5
1.5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Math_Int(t *testing.T) {
	rsl := `
print(1 + 3)
print(3 / 2)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `4
1.5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Math_ErrorsOnIntIntDivisionByZero(t *testing.T) {
	rsl := `
a = 1 / 0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L2:9

  a = 1 / 0
          ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_ErrorsOnFloatIntDivisionByZero(t *testing.T) {
	rsl := `
a = 1.0 / 0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L2:11

  a = 1.0 / 0
            ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_ErrorsOnIntFloatDivisionByZero(t *testing.T) {
	rsl := `
a = 1 / 0.0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L2:9

  a = 1 / 0.0
          ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_ErrorsOnFloatFloatDivisionByZero(t *testing.T) {
	rsl := `
a = 1.0 / 0.0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L2:11

  a = 1.0 / 0.0
            ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_CompoundDivideByZeroIntIntErrors(t *testing.T) {
	rsl := `
a = 1
a /= 0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L3:6

  a /= 0
       ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_CompoundDivideByZeroIntFloatErrors(t *testing.T) {
	rsl := `
a = 1
a /= 0.0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L3:6

  a /= 0.0
       ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_CompoundDivideByZeroFloatIntErrors(t *testing.T) {
	rsl := `
a = 1.0
a /= 0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L3:6

  a /= 0
       ^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_CompoundDivideByZeroFloatFloatErrors(t *testing.T) {
	rsl := `
a = 1.0
a /= 0.0
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L3:6

  a /= 0.0
       ^^^ Divisor was 0, cannot divide by 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Math_CanHaveManyPluses(t *testing.T) {
	rsl := `
a = 1 +++++++++ +2
print(a)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Math_CanHaveManyMinuses(t *testing.T) {
	rsl := `
a = 1 + -------2
print(a)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "-1\n")
	assertNoErrors(t)
	resetTestState()
}
