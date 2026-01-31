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
	assertErrorContains(t, 1, "RAD30001", "is not compatible with expected type")
}

func Test_Func_Max_Negative(t *testing.T) {
	script := `
print(max([-1, -2.2, -3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1\n")
	assertNoErrors(t)
}

func Test_Func_Max_Variadic_Basic(t *testing.T) {
	script := `
print(max(1, 3, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Func_Max_Variadic_MixedIntFloat(t *testing.T) {
	script := `
print(max(1, 3.5, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3.5\n")
	assertNoErrors(t)
}

func Test_Func_Max_Variadic_SingleNumber(t *testing.T) {
	script := `
print(max(5))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
}

func Test_Func_Max_Variadic_Negative(t *testing.T) {
	script := `
print(max(-1, -2.5, 3))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Func_Max_SingleElementList(t *testing.T) {
	script := `
print(max([5]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
}

func Test_Func_Max_EmptyListError(t *testing.T) {
	script := `
print(max([]))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20018", "Cannot find max of empty list")
}

func Test_Func_Max_NoArgsError(t *testing.T) {
	script := `
print(max())
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20018", "Cannot find max of empty list")
}

func Test_Func_Max_MultipleListsError(t *testing.T) {
	script := `
print(max([1, 2], [3, 4]))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20012", "max() with multiple arguments requires numbers, not lists")
}
