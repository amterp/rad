package testing

import "testing"

func Test_Func_Sum_Ints(t *testing.T) {
	script := `
a = [1, 2, 3]
print(sum(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "6\n")
	assertNoErrors(t)
}

func Test_Func_Sum_Mix(t *testing.T) {
	script := `
a = [1, 2.2, 3]
print(sum(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "6.2\n")
	assertNoErrors(t)
}

func Test_Func_Sum_ErrorsForNonNumElements(t *testing.T) {
	script := `
a = [1, "ab", 3]
print(sum(a))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:11

  print(sum(a))
            ^
            Value '[ 1, "ab", 3 ]' (list) is not compatible with expected type 'float[]'
`
	assertError(t, 1, expected)
}
