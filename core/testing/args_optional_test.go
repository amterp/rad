package testing

import (
	"testing"
)

func Test_Args_Optional(t *testing.T) {
	script := `
args:
    name str
    age int
    role str?
    year int?

print(name, age, role, year, sep="|")
`
	setupAndRunCode(t, script, "hey", "30", "--color=never")
	expected := `hey|30|null|null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Optional_RelationalRequiresMet(t *testing.T) {
	script := `
args:
    name str
    age int?
	name requires age

print(name, age, sep="|")
`
	setupAndRunCode(t, script, "hey", "30", "--color=never")
	expected := `hey|30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Args_Optional_RelationalRequiresNotMet(t *testing.T) {
	script := `
args:
    name str
    age int?
	name requires age

print(name, age, sep="|")
`
	setupAndRunCode(t, script, "hey", "--color=never")
	expected := `Invalid args: 'name' requires 'age', but 'age' was not set

Usage:
  TestCase <name> [age] [OPTIONS]

Script args:
      --name str   Requires: age
      --age int

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}
