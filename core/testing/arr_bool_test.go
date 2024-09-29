package testing

import "testing"

func TestBoolArrays(t *testing.T) {
	rsl := `
a bool[] = [true, true, false]
print(a)
print(join(a, "-"))
//print(a + [true]) // todo implement
//print(a + true)
`
	setupAndRunCode(t, rsl)
	expected := `[true, true, false]
true-true-false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
