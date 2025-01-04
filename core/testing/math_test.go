package testing

import "testing"

func Test_Math_Float(t *testing.T) {
	rsl := `
print(1.2 + 2.3)
print(3.0 / 2.0)
print(3.0 / 2)
`
	setupAndRunCode(t, rsl)
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
	setupAndRunCode(t, rsl)
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
print(a)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/7 on '/': Cannot divide by 0\n")
	resetTestState()
}

func Test_Math_ErrorsOnFloatIntDivisionByZero(t *testing.T) {
	rsl := `
a = 1.0 / 0
print(a)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/9 on '/': Cannot divide by 0\n")
	resetTestState()
}

func Test_Math_ErrorsOnIntFloatDivisionByZero(t *testing.T) {
	rsl := `
a = 1 / 0.0
print(a)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/7 on '/': Cannot divide by 0\n")
	resetTestState()
}

func Test_Math_ErrorsOnFloatFloatDivisionByZero(t *testing.T) {
	rsl := `
a = 1.0 / 0.0
print(a)
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/9 on '/': Cannot divide by 0\n")
	resetTestState()
}

func Test_Math_CompoundDivideByZeroIntIntErrors(t *testing.T) {
	rsl := `
a = 1
a /= 0
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/4 on '/=': Cannot divide by 0\n")
	resetTestState()
}

func Test_Math_CompoundDivideByZeroIntFloatErrors(t *testing.T) {
	rsl := `
a = 1
a /= 0.0
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/4 on '/=': Cannot divide by 0\n")
	resetTestState()
}

func Test_Math_CompoundDivideByZeroFloatIntErrors(t *testing.T) {
	rsl := `
a = 1.0
a /= 0
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/4 on '/=': Cannot divide by 0\n")
	resetTestState()
}

func Test_Math_CompoundDivideByZeroFloatFloatErrors(t *testing.T) {
	rsl := `
a = 1.0
a /= 0.0
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/4 on '/=': Cannot divide by 0\n")
	resetTestState()
}
