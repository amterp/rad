package testing

import "testing"

func TestBoolArrays_General(t *testing.T) {
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

// todo uncomment when += operations implemented for bool arrays
//func TestBoolArrays_CanModify(t *testing.T) {
//	rsl := `
//a bool[] = [true, true, false]
//a += [true]
//print(a)
//`
//	setupAndRunCode(t, rsl)
//	expected := `[true, true, false, true]
//`
//	assertOnlyOutput(t, stdOutBuffer, expected)
//	assertNoErrors(t)
//	resetTestState()
//}
