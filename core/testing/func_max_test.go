package testing

import "testing"

func Test_Func_Max_Ints(t *testing.T) {
	script := `
print(max([1, 2, 3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Func_Max_Mix(t *testing.T) {
	script := `
print(max([1, 2.2, 3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Func_Max_ErrorsForNonNumElements(t *testing.T) {
	script := `
print(max([1, "ab", 3]))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:11

  print(max([1, "ab", 3]))
            ^^^^^^^^^^^^ max() requires a list of numbers, got "str" at index 1
`
	assertError(t, 1, expected)
}

func Test_Func_Max_Negative(t *testing.T) {
	script := `
print(max([-1, -2.2, -3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1\n")
	assertNoErrors(t)
}
