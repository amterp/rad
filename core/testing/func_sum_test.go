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
	assertErrorContains(t, 1, "RAD30001", "Value '[ 1, \"ab\", 3 ]' (list) is not compatible with expected type 'float[]'")
}

func Test_Func_Sum_ReturnsInt_WhenAllInts(t *testing.T) {
	script := `
print(type_of(sum([1, 2, 3])))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "int\n")
	assertNoErrors(t)
}

func Test_Func_Sum_ReturnsFloat_WhenAnyFloat(t *testing.T) {
	script := `
print(type_of(sum([1, 2.0, 3])))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "float\n")
	assertNoErrors(t)
}
