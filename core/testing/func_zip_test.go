package testing

import "testing"

func Test_Func_Zip_NoArgs(t *testing.T) {
	script := `
print(zip())
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ ]\n")
	assertNoErrors(t)
}

func Test_Func_Zip_OneList(t *testing.T) {
	script := `
print(zip([1, 2, 3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ [ 1 ], [ 2 ], [ 3 ] ]\n")
	assertNoErrors(t)
}

func Test_Func_Zip_TwoLists(t *testing.T) {
	script := `
print(zip([1, 2, 3], ["a", "b", "c"]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ], [ 3, "c" ] ]`+"\n")
	assertNoErrors(t)
}

func Test_Func_Zip_FourLists(t *testing.T) {
	script := `
print(zip([1, 2, 3], ["a", "b", "c"], [4, 5, 6], ["d", "e", "f"]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a", 4, "d" ], [ 2, "b", 5, "e" ], [ 3, "c", 6, "f" ] ]`+"\n")
	assertNoErrors(t)
}

func Test_Func_Zip_StopsOnShorter(t *testing.T) {
	script := `
print(zip([1, 2, 3, 4], ["a", "b"]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ] ]`+"\n")
	assertNoErrors(t)
}

func Test_Func_Zip_FillsToLongerIfProvided(t *testing.T) {
	script := `
print(zip([1, 2, 3, 4], ["a", "b"], fill="-"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ], [ 3, "-" ], [ 4, "-" ] ]`+"\n")
	assertNoErrors(t)
}

func Test_Func_Zip_DoesNotErrorIfStrictAndListsSameLength(t *testing.T) {
	script := `
print(zip([1, 2, 3, 4], ["a", "b", "c", "d"], strict=true))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ], [ 3, "c" ], [ 4, "d" ] ]`+"\n")
	assertNoErrors(t)
}

func Test_Func_Zip_ErrorsIfStrictAndListsNotSameLength(t *testing.T) {
	script := `
print(zip([1, 2, 3, 4], ["a", "b"], strict=true))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20015", "Strict mode enabled: all lists must have the same length, but got 4 and 2")
}

func Test_Func_Zip_ErrorsIfStrictAndFillProvided(t *testing.T) {
	script := `
print(zip([1, 2, 3, 4], ["a", "b", "c", "d"], strict=true, fill="-"))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20014", "Cannot enable 'strict' with 'fill' specified")
}

func Test_Func_Zip_EmptyLists(t *testing.T) {
	script := `
print(zip([], [], []))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `[ ]`+"\n")
	assertNoErrors(t)
}
