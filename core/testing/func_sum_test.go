package testing

import "testing"

func Test_Func_Sum_Ints(t *testing.T) {
	rsl := `
a = [1, 2, 3]
print(sum(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "6\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Sum_Mix(t *testing.T) {
	rsl := `
a = [1, 2.2, 3]
print(sum(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "6.2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Sum_ErrorsForNonNumElements(t *testing.T) {
	rsl := `
a = [1, "ab", 3]
print(sum(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L3:11

  print(sum(a))
            ^ sum() requires a list of numbers, got "string" at index 1
`
	assertError(t, 1, expected)
	resetTestState()
}
