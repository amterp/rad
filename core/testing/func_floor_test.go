package testing

import "testing"

func Test_Func_Floor_Ints(t *testing.T) {
	script := `
print(floor(1))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Negative_Ints(t *testing.T) {
	script := `
print(floor(-2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Floats(t *testing.T) {
	script := `
print(floor(2.234))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Negative_Floats(t *testing.T) {
	script := `
print(floor(-2.234))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-3\n")
	assertNoErrors(t)
}

func Test_Func_Floor_Errors_With_String(t *testing.T) {
	script := `
print(floor("ab"))
	`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:13

  print(floor("ab"))
              ^^^^ Value '"ab"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}
