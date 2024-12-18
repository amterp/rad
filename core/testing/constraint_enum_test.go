package testing

import "testing"

func Test_Constraint_Enum_Valid(t *testing.T) {
	rsl := `
args:
	name string
	enum name ["alice", "bob", "charlie"]
print("Hi", name)
`
	setupAndRunCode(t, rsl, "alice")
	assertOnlyOutput(t, stdOutBuffer, "Hi alice\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Constraint_Enum_ErrorsOnInvalid(t *testing.T) {
	rsl := `
args:
	name string
	enum name ["alice", "bob", "charlie"]
print("Hi", name)
`
	setupAndRunCode(t, rsl, "david", "--NO-COLOR")
	expected := `Invalid 'name' value: david (valid values: alice, bob, charlie)
Usage:
  test <name>

Script args:
      --name string    Valid values: [alice, bob, charlie]

` + globalFlagHelp
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Constraint_Enum_ErrorsIfNonStringEnum(t *testing.T) {
	rsl := `
args:
	name string
	enum name ["alice", 2]
print("Hi", name)
`
	setupAndRunCode(t, rsl, "david", "--NO-COLOR")
	assertError(t, 1, "RslError at L4/23 on '2': Expected string literal, got int\n")
	resetTestState()
}
