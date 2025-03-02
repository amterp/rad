package testing

import "testing"

func Test_Func_Zip_NoArgs(t *testing.T) {
	rsl := `
print(zip())
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "[ ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Zip_OneList(t *testing.T) {
	rsl := `
print(zip([1, 2, 3]))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "[ [ 1 ], [ 2 ], [ 3 ] ]\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Zip_TwoLists(t *testing.T) {
	rsl := `
print(zip([1, 2, 3], ["a", "b", "c"]))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ], [ 3, "c" ] ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Zip_FourLists(t *testing.T) {
	rsl := `
print(zip([1, 2, 3], ["a", "b", "c"], [4, 5, 6], ["d", "e", "f"]))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a", 4, "d" ], [ 2, "b", 5, "e" ], [ 3, "c", 6, "f" ] ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Zip_StopsOnShorter(t *testing.T) {
	rsl := `
print(zip([1, 2, 3, 4], ["a", "b"]))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ] ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Zip_FillsToLongerIfProvided(t *testing.T) {
	rsl := `
print(zip([1, 2, 3, 4], ["a", "b"], fill="-"))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ], [ 3, "-" ], [ 4, "-" ] ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Zip_DoesNotErrorIfStrictAndListsSameLength(t *testing.T) {
	rsl := `
print(zip([1, 2, 3, 4], ["a", "b", "c", "d"], strict=true))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, `[ [ 1, "a" ], [ 2, "b" ], [ 3, "c" ], [ 4, "d" ] ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Zip_ErrorsIfStrictAndListsNotSameLength(t *testing.T) {
	rsl := `
print(zip([1, 2, 3, 4], ["a", "b"], strict=true))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L2:7

  print(zip([1, 2, 3, 4], ["a", "b"], strict=true))
        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        Strict mode enabled: all lists must have the same length, but got 4 and 2
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Func_Zip_ErrorsIfStrictAndFillProvided(t *testing.T) {
	rsl := `
print(zip([1, 2, 3, 4], ["a", "b", "c", "d"], strict=true, fill="-"))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L2:7

  print(zip([1, 2, 3, 4], ["a", "b", "c", "d"], strict=true, fill="-"))
        ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
        Cannot specify both 'strict' and 'fill' named arguments
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Func_Zip_EmptyLists(t *testing.T) {
	rsl := `
print(zip([], [], []))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, `[ ]`+"\n")
	assertNoErrors(t)
	resetTestState()
}
