package testing

import "testing"

func Test_Func_Min_Ints(t *testing.T) {
	rsl := `
a = [1, 2, 3]
print(min(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Min_Mix(t *testing.T) {
	rsl := `
a = [1, 2.2, 3]
print(min(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Min_ErrorsForNonNumElements(t *testing.T) {
	rsl := `
a = [1, "ab", 3]
print(min(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:11

  print(min(a))
            ^ min() requires a list of numbers, got "string" at index 1
`
	assertError(t, 1, expected)
	resetTestState()
}
