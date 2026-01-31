package testing

import "testing"

func Test_Func_Float_PassthroughForFloat(t *testing.T) {
	script := `
a = float(10.2)
print(a)
print(type_of(a))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10.2
float
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Float_ConvertsInt(t *testing.T) {
	script := `
a = float(10)
print(a)
print(type_of(a))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10
float
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Float_Bool(t *testing.T) {
	script := `
a = float(true)
b = float(false)
print(a, b)
print(type_of(a), type_of(b))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1 0
float float
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Float_ErrorsOnMap(t *testing.T) {
	script := `
float({})
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20016", "Cannot cast \"map\" to float")
}

func Test_Func_Float_ErrorsOnStringWithDetails(t *testing.T) {
	script := `
float("10")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20016", "Cannot cast string to float. Did you mean to use 'parse_float' to parse the given string?")
}
