package testing

import "testing"

func Test_Constraint_Regex_Help(t *testing.T) {
	script := `
args:
	name str
	name regex "[A-Z][a-z]*"
print("Hi", name)
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str   Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Constraint_Regex_Help_WithEnum(t *testing.T) {
	script := `
args:
	name str
	name regex "[A-Z][a-z]*"
	name enum ["Alice", "Bob"]
print("Hi", name)
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str   Valid values: [Alice, Bob]. Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Constraint_Regex_Valid(t *testing.T) {
	script := `
args:
	name str
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
	name str
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
	name str
	name regex "[A-Z][a-z]*"
`
	setupAndRunCode(t, script, "alice", "--color=never")
	expected := `Invalid 'name' value: alice (must match regex: [A-Z][a-z]*)

Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str   Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Constraint_RegexAndEnum_InvalidInput(t *testing.T) {
	script := `
args:
	name str
	name regex "[A-Z][a-z]*"
	name enum ["Alice", "Bob"]
`
	setupAndRunCode(t, script, "Charlie", "--color=never")
	expected := `Invalid 'name' value: Charlie (valid values: Alice, Bob)

Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str   Valid values: [Alice, Bob]. Regex: [A-Z][a-z]*

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

// not clear color here should be present or not. comes down to if we eagerly parse flags we can, or not.
func Test_Constraint_Regex_InvalidRegex(t *testing.T) {
	script := `
args:
	name str
	name regex "+"
`
	setupAndRunCode(t, script, "--color=never")
	expected := "\x1b[33mError at L4:2\n\n\x1b[0m  \tname regex \"+\"\n   \x1b[31m^^^^^^^^^^^^^^\x1b[0m\n   \x1b[31mInvalid regex '+': error parsing regexp: missing argument to repetition operator: `+`\x1b[0m\n"
	assertError(t, 1, expected)
}
