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
