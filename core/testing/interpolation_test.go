package testing

import "testing"

func TestStringInterpolation_String(t *testing.T) {
	rsl := `
var = "alice"
print("hello, {var}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hello, alice\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStringInterpolation_Int(t *testing.T) {
	rsl := `
var = 42
print("hello, {var}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hello, 42\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStringInterpolation_Float(t *testing.T) {
	rsl := `
var = 12.5
print("hello, {var}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hello, 12.5\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStringInterpolation_Bool(t *testing.T) {
	rsl := `
var = true
print("hello, {var}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hello, true\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStringInterpolation_List(t *testing.T) {
	rsl := `
var = ["alice", 42]
print("hello, {var}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hello, [alice, 42]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStringInterpolation_Map(t *testing.T) {
	rsl := `
var = { "name": "alice", "age": 42 }
print("hello, {var}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hello, { name: alice, age: 42 }\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStringInterpolation_ErrorsIfUnknownVariable(t *testing.T) {
	rsl := `
print("hello, {var}")
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/20 on '\"hello, {var}\"': Undefined variable referenced: var\n")
	resetTestState()
}

func TestStringInterpolation_CanEscapeFirst(t *testing.T) {
	rsl := `
print("hello, \{var}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hello, {var}\n")
	assertNoErrors(t)
	resetTestState()
}

// todo this should either be handled better or just pass
//func TestStringInterpolation_CanEscapeSecond(t *testing.T) {
//	rsl := `
//print("hello, {var\}")
//`
//	setupAndRunCode(t, rsl)
//	assertOnlyOutput(t, stdOutBuffer, "hello, {var}\n")
//	assertNoErrors(t)
//	resetTestState()
//}
