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
	assertErrorContains(t, 1, "RAD30001", "is not compatible with expected type")
}

func Test_Func_Min_Negative(t *testing.T) {
	script := `
print(min([-1, -2.2, -3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-3\n")
	assertNoErrors(t)
}

func Test_Func_Min_Variadic_Basic(t *testing.T) {
	script := `
print(min(3, 1, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Min_Variadic_MixedIntFloat(t *testing.T) {
	script := `
print(min(3, 1.5, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1.5\n")
	assertNoErrors(t)
}

func Test_Func_Min_Variadic_SingleNumber(t *testing.T) {
	script := `
print(min(5))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
}

func Test_Func_Min_Variadic_Negative(t *testing.T) {
	script := `
print(min(-1, -2.5, 3))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-2.5\n")
	assertNoErrors(t)
}

func Test_Func_Min_SingleElementList(t *testing.T) {
	script := `
print(min([5]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
}

func Test_Func_Min_EmptyListError(t *testing.T) {
	script := `
print(min([]))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20018", "Cannot find min of empty list")
}

func Test_Func_Min_NoArgsError(t *testing.T) {
	script := `
print(min())
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20018", "Cannot find min of empty list")
}

func Test_Func_Min_MultipleListsError(t *testing.T) {
	script := `
print(min([1, 2], [3, 4]))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20012", "min() with multiple arguments requires numbers, not lists")
}
