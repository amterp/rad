package testing

import "testing"

func Test_Func_Min_Ints(t *testing.T) {
	script := `
print(min([1, 2, 3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Min_Mix(t *testing.T) {
	script := `
print(min([1, 2.2, 3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Min_ErrorsForNonNumElements(t *testing.T) {
	script := `
print(min([1, "ab", 3]))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:11

  print(min([1, "ab", 3]))
            ^^^^^^^^^^^^ min() requires a list of numbers, got "string" at index 1
`
	assertError(t, 1, expected)
}

func Test_Func_Min_Negative(t *testing.T) {
	script := `
print(min([-1, -2.2, -3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-3\n")
	assertNoErrors(t)
}
