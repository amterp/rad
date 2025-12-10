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
            ^^^^^^^^^^^^
            Value '[ 1, "ab", 3 ]' (list) is not compatible with expected type 'float|float[]'
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
	expected := `Error at L2:7

  print(min([]))
        ^^^^^^^ Cannot find min of empty list (RAD20018)
`
	assertError(t, 1, expected)
}

func Test_Func_Min_NoArgsError(t *testing.T) {
	script := `
print(min())
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:7

  print(min())
        ^^^^^ Cannot find min of empty list (RAD20018)
`
	assertError(t, 1, expected)
}

func Test_Func_Min_MultipleListsError(t *testing.T) {
	script := `
print(min([1, 2], [3, 4]))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:7

  print(min([1, 2], [3, 4]))
        ^^^^^^^^^^^^^^^^^^^
        min() with multiple arguments requires numbers, not lists. Use min([...]) for a single list (RAD20012)
`
	assertError(t, 1, expected)
}
