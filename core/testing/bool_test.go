package testing

import "testing"

func TestBool_And(t *testing.T) {
	script := `
print(true and true)
print(true and false)
print(false and true)
print(false and false)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\nfalse\nfalse\n")
	assertNoErrors(t)
}

func TestBool_Or(t *testing.T) {
	script := `
print(true or true)
print(true or false)
print(false or true)
print(false or false)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\ntrue\ntrue\nfalse\n")
	assertNoErrors(t)
}

func TestBool_Not(t *testing.T) {
	script := `
print(true)
print(not true)
print(false)
print(not false)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\nfalse\ntrue\n")
	assertNoErrors(t)
}

func TestBool_Equality(t *testing.T) {
	script := `
print(true == true)
print(true == false)
print(false == true)
print(false == false)

print(true != true)
print(true != false)
print(false != true)
print(false != false)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\nfalse\nfalse\ntrue\nfalse\ntrue\ntrue\nfalse\n")
	assertNoErrors(t)
}
