package testing

import "testing"

func Test_Func_Floor_Ints(t *testing.T) {
	rsl := `
a = 1
print(floor(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Floor_Negative_Ints(t *testing.T) {
	rsl := `
a = -2
print(floor(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Floor_Floats(t *testing.T) {
	rsl := `
a = 2.234
print(floor(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Floor_Negative_Floats(t *testing.T) {
	rsl := `
a = -2.234
print(floor(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Floor_Errors_With_String(t *testing.T) {
	rsl := `
b = "ab"
print(floor(b))
	`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:13

  print(floor(b))
              ^
              Got "string" as the 1st argument of floor(), but must be: float or int
`
	assertError(t, 1, expected)
	resetTestState()
}
