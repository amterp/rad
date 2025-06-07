package testing

import (
	"testing"
)

func Test_Error_Creation(t *testing.T) {
	script := `
err = error("test error message")
print(err)
`
	setupAndRunCode(t, script)
	assertOnlyOutput(t, stdOutBuffer, "test error message\n")
	assertNoErrors(t)
}

func Test_Error_Equality(t *testing.T) {
	script := `
err1 = error("error1")
err2 = error("error1")
err3 = error("error2")

print(err1 == err2)  // true - same message
print(err1 != err2)  // false - same message
print(err1 == err3)  // false - different message
print(err1 != err3)  // true - different message
`
	setupAndRunCode(t, script)
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\nfalse\ntrue\n")
	assertNoErrors(t)
}

func Test_Error_StringOperations(t *testing.T) {
	script := `
err = error("error message")
str = "string"

// Test concatenation
print(err + str)
print(str + err)

// Test equality with string
print(err == "error message")
print(err != "error message")
print(err == "different")
print(err != "different")

// Test in/not in operations
print(err in "This is an error message")
print(err not in "This is an error message")
print(err in "No match here")
print(err not in "No match here")
`
	setupAndRunCode(t, script)
	expected := `error messagestring
stringerror message
true
false
false
true
true
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Error_NumericOperations(t *testing.T) {
	script := `
err = error("error")

// Test concatenation with numbers
print(err + 123)
print(err + 3.14)
`
	setupAndRunCode(t, script)
	expected := `error123
error3.14
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Error_BooleanOperations(t *testing.T) {
	script := `
err = error("error")

// Test concatenation with booleans
print(err + true)
print(err + false)
`
	setupAndRunCode(t, script)
	expected := `errortrue
errorfalse
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Error_CollectionOperations(t *testing.T) {
	script := `
err1 = error("item1")
err2 = error("item2")
err3 = error("item3")
err4 = error("item4")

// Create list and map with error objects
list = [err1, err2, err3]
map = {err1: 1, err2: 2, err3: 3}

// Test in/not in operations with list
print(err1 in list)
print(err2 in list)
print(err3 in list)
print(err4 in list)
print(err4 not in list)

// Test in/not in operations with map
print(err1 in map)
print(err2 in map)
print(err3 in map)
print(err4 in map)
print(err4 not in map)
`
	setupAndRunCode(t, script)
	expected := `true
true
true
false
true
true
true
true
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Error_ErrorOperations(t *testing.T) {
	script := `
err1 = error("error1")
err2 = error("error2")

print(err1 + err2)

print(err1 == error("error1"))
print(err1 != error("error1"))
print(err1 == err2)
print(err1 != err2)
`
	setupAndRunCode(t, script)
	expected := `error1error2
true
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Error_Printing(t *testing.T) {
	script := `
err = error("test error message")
print(err)
`
	setupAndRunCode(t, script)
	expected := `test error message
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
