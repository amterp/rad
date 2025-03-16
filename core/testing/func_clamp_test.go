package testing

import "testing"

func Test_Func_Clamp_Ints(t *testing.T) {
	rsl := `
print(clamp(1, 0, 2))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Clamp_Mix(t *testing.T) {
	rsl := `
print(clamp(2.2, 1.2, 2))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Func_Clamp_ErrorsForLessThan3Elements(t *testing.T) {
	rsl := `
print(clamp(1, 2))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:7

  print(clamp(1, 2))
        ^^^^^^^^^^^ clamp() requires at least 3 arguments, but got 2
`
	assertError(t, 1, expected)
}

func Test_Func_Clamp_ErrorsForNonNumElements(t *testing.T) {
	rsl := `
print(clamp(1, "ab", 2))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:16

  print(clamp(1, "ab", 2))
                 ^^^^
                 Got "string" as the 2nd argument of clamp(), but must be: float or int
`
	assertError(t, 1, expected)
}

func Test_Func_Clamp_Negative(t *testing.T) {
	rsl := `
print(clamp(-2.2, -1.2, 2))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1.2\n")
	assertNoErrors(t)
}
