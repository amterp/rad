package testing

import "testing"

func Test_Func_Float_PassthroughForFloat(t *testing.T) {
	rsl := `
a = float(10.2)
print(a)
print(type_of(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `10.2
float
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Float_ConvertsInt(t *testing.T) {
	rsl := `
a = float(10)
print(a)
print(type_of(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `10
float
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Float_Bool(t *testing.T) {
	rsl := `
a = float(true)
b = float(false)
print(a, b)
print(type_of(a), type_of(b))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `1 0
float float
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Float_ErrorsOnMap(t *testing.T) {
	rsl := `
float({})
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:7

  float({})
        ^^ Cannot cast "map" to float
`
	assertError(t, 1, expected)
}

func Test_Func_Float_ErrorsOnStringWithDetails(t *testing.T) {
	rsl := `
float("10")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:7

  float("10")
        ^^^^
        Cannot cast string to float. Did you mean to use 'parse_float' to parse the given string?
`
	assertError(t, 1, expected)
}
