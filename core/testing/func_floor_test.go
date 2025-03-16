package testing

import "testing"

func Test_Func_Floor_Ints(t *testing.T) {
	rsl := `
print(floor(1))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Negative_Ints(t *testing.T) {
	rsl := `
print(floor(-2))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Floats(t *testing.T) {
	rsl := `
print(floor(2.234))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Negative_Floats(t *testing.T) {
	rsl := `
print(floor(-2.234))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-3\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Errors_With_String(t *testing.T) {
	rsl := `
print(floor("ab"))
	`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:13

  print(floor("ab"))
              ^^^^
              Got "string" as the 1st argument of floor(), but must be: float or int
`
	assertError(t, 1, expected)
}
