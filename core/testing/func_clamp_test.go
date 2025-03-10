package testing

import "testing"

func Test_Func_Clamp_Ints(t *testing.T) {
	rsl := `
a = 1
b = 0
c = 2
print(clamp(a, b, c))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Clamp_Mix(t *testing.T) {
	rsl := `
a = 2.2
b = 1.2
c = 2
print(clamp(a, b, c))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Clamp_ErrorsForLessThan3Elements(t *testing.T) {
	rsl := `
a = 1
b = 2
print(clamp(a, b))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L4:7

  print(clamp(a, b))
        ^^^^^^^^^^^ clamp() requires at least 3 arguments, but got 2
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Func_Clamp_ErrorsForNonNumElements(t *testing.T) {
	rsl := `
a = 1
b = "ab"
c = 2
print(clamp(a, b, c))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L5:16

  print(clamp(a, b, c))
                 ^
                 Got "string" as the 2nd argument of clamp(), but must be: float or int
`
	assertError(t, 1, expected)
	resetTestState()
}
