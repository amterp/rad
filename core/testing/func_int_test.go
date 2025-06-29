package testing

import "testing"

func Test_Func_Int_PassthroughForInt(t *testing.T) {
	script := `
a = int(10)
print(a)
print(type_of(a))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10
int
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Int_FloorsFloat(t *testing.T) {
	script := `
a = int(10.7)
print(a)
print(type_of(a))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10
int
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Int_Bool(t *testing.T) {
	script := `
a = int(true)
b = int(false)
print(a, b)
print(type_of(a), type_of(b))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1 0
int int
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Int_ErrorsOnMap(t *testing.T) {
	script := `
int({})
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  int({})
  ^^^^^^^ Cannot cast "map" to int (RAD20016)
`
	assertError(t, 1, expected)
}

func Test_Func_Int_ErrorsOnStringWithDetails(t *testing.T) {
	script := `
int("10")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  int("10")
  ^^^^^^^^^
  Cannot cast string to int. Did you mean to use 'parse_int' to parse the given string? (RAD20016)
`
	assertError(t, 1, expected)
}
