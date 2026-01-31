package testing

import "testing"

func Test_Func_Clamp_Ints(t *testing.T) {
	script := `
print(clamp(1, 0, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Func_Clamp_Mix(t *testing.T) {
	script := `
print(clamp(2.2, 1.2, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Func_Clamp_ErrorsForLessThan3Elements(t *testing.T) {
	script := `
print(clamp(1, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30007", "Missing required argument 'max'")
}

func Test_Func_Clamp_ErrorsForNonNumElements(t *testing.T) {
	script := `
print(clamp(1, "ab", 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30001", "Value '\"ab\"' (str) is not compatible with expected type 'float'")
}

func Test_Func_Clamp_Negative(t *testing.T) {
	script := `
print(clamp(-2.2, -1.2, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1.2\n")
	assertNoErrors(t)
}
