package testing

import "testing"

func Test_Constraint_Regex_Help(t *testing.T) {
	script := `
args:
	name string
	name regex "[A-Z][a-z]*"
print("Hi", name)
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  <name> [OPTIONS]

Script args:
      --name string   Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Constraint_Regex_Help_WithEnum(t *testing.T) {
	script := `
args:
	name string
	name regex "[A-Z][a-z]*"
	name enum ["Alice", "Bob"]
print("Hi", name)
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  <name> [OPTIONS]

Script args:
      --name string   Valid values: [Alice, Bob]. Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Constraint_Regex_Valid(t *testing.T) {
	script := `
args:
	name string
	name regex "[A-Z][a-z]*"
print("Hi", name)
`
	setupAndRunCode(t, script, "Alice")
	assertOnlyOutput(t, stdOutBuffer, "Hi Alice\n")
	assertNoErrors(t)
}

func Test_Constraint_RegexAndEnum_Valid(t *testing.T) {
	script := `
args:
	name string
	name regex "[A-Z][a-z]*"
	name enum ["Alice", "Bob"]
print("Hi", name)
`
	setupAndRunCode(t, script, "Alice")
	assertOnlyOutput(t, stdOutBuffer, "Hi Alice\n")
	assertNoErrors(t)
}

func Test_Constraint_Regex_InvalidInput(t *testing.T) {
	script := `
args:
	name string
	name regex "[A-Z][a-z]*"
`
	setupAndRunCode(t, script, "alice", "--color=never")
	expected := `Invalid 'name' value: alice (must match regex: [A-Z][a-z]*)

Usage:
  <name> [OPTIONS]

Script args:
      --name string   Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Constraint_RegexAndEnum_InvalidInput(t *testing.T) {
	script := `
args:
	name string
	name regex "[A-Z][a-z]*"
	name enum ["Alice", "Bob"]
`
	setupAndRunCode(t, script, "Charlie", "--color=never")
	expected := `Invalid 'name' value: Charlie (valid values: Alice, Bob)

Usage:
  <name> [OPTIONS]

Script args:
      --name string   Valid values: [Alice, Bob]. Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Constraint_Regex_InvalidRegex(t *testing.T) {
	script := `
args:
	name string
	name regex "+"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L4:2

  	name regex "+"
   ^^^^^^^^^^^^^^
   Invalid regex '+': error parsing regexp: missing argument to repetition operator: ` + "`+`\n"
	assertError(t, 1, expected)
}
