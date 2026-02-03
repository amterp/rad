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
	assertErrorContains(t, 1, "RAD30001", "Value '\"ab\"' (str) is not compatible with expected type 'int|float'")
}

func Test_Func_Clamp_Negative(t *testing.T) {
	script := `
print(clamp(-2.2, -1.2, 2))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-1.2\n")
	assertNoErrors(t)
}

func Test_Func_Clamp_ReturnsInt_WhenAllInts(t *testing.T) {
	script := `
print(type_of(clamp(5, 1, 10)))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "int\n")
	assertNoErrors(t)
}

func Test_Func_Clamp_ReturnsFloat_WhenAnyFloat(t *testing.T) {
	script := `
print(type_of(clamp(5, 1.0, 10)))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "float\n")
	assertNoErrors(t)
}

func Test_Func_Clamp_CanBeUsedForListIndexing(t *testing.T) {
	script := `
mylist = [10, 20, 30, 40, 50]
myidx = 7
print(mylist[clamp(myidx, 0, 4)])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "50\n")
	assertNoErrors(t)
}
