package testing

import "testing"

func Test_Func_Max_Ints(t *testing.T) {
	rsl := `
a = [1, 2, 3]
print(max(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Max_Mix(t *testing.T) {
	rsl := `
a = [1, 2.2, 3]
print(max(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Max_ErrorsForNonNumElements(t *testing.T) {
	rsl := `
a = [1, "ab", 3]
print(max(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:11

  print(max(a))
            ^ max() requires a list of numbers, got "string" at index 1
`
	assertError(t, 1, expected)
	resetTestState()
}
