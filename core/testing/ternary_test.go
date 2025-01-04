package testing

import "testing"

func Test_Ternary_True(t *testing.T) {
	rsl := `
a = "alice"
print(true ? a : "bob")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "alice\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_False(t *testing.T) {
	rsl := `
a = "alice"
print(false ? a : "bob")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "bob\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_Nested(t *testing.T) {
	rsl := `
print(true ? (false ? "bob" : "charlie") : "alice")
print(true ? false ? "bob" : "charlie" : "alice")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "charlie\ncharlie\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_Complex(t *testing.T) {
	rsl := `
a = 5
b = "alice"
c = "charlie"
print((c[0] == 'c' ? c : b)[(len(b) > 3 ? 1 : 2):5])
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "harl\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_Truthy(t *testing.T) {
	rsl := `
a = "not empty"
print(a ? a : "empty")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_Falsy(t *testing.T) {
	rsl := `
a = ""
print(a ? a : "empty")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_CanDefineAcrossLines1(t *testing.T) {
	rsl := `
a = "blah"
	? "not empty"
	: "empty"
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_CanDefineAcrossLines2(t *testing.T) {
	rsl := `
a = "blah" ? "not empty"
	: "empty"
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_CanDefineAcrossLines3(t *testing.T) {
	rsl := `
a = "blah" 
	? "not empty" : "empty"
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_CanDefineAcrossLines4(t *testing.T) {
	rsl := `
a = "blah" 
	? 
	"not empty" : "empty"
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_CanDefineAcrossLines5(t *testing.T) {
	rsl := `
a = "blah"
	? 
	"not empty" : 
	"empty"
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_CanDefineAcrossLines6BigSpace(t *testing.T) {
	rsl := `
a = "blah"


	?


	"not empty" 			: 


	"empty"
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "not empty\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Ternary_FailsIfWhitespaceAbused(t *testing.T) {
	rsl := `
if true:
	a = "blah" 
? "not empty" : "empty"
	print("one")

print("two")
`
	setupAndRunCode(t, rsl)
	// this specific error is not ideal, we can (and probably should) improve it
	assertError(t, 1, "RslError at L5/2 on '\t': Expected expression\n")
	resetTestState()
}
