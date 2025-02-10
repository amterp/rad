package testing

import "testing"

func TestSwitch_CanSelectCaseBasedOnKeys(t *testing.T) {
	t.Skip("TODO RAD-141") // TODO RAD-141
	rsl := `
name = "alice"
result1 = switch name:
	case "alice": "ALICE"
	case "bob": "BOB"
	case "charlie": "CHARLIE"
print(result1)

name = "bob"
result2 = switch name:
	case "alice": "ALICE"
	case "bob": "BOB"
	case "charlie": "CHARLIE"
print(result2)

name = "charlie"
result3 = switch name:
	case "alice": "ALICE"
	case "bob": "BOB"
	case "charlie": "CHARLIE"
print(result3)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\nBOB\nCHARLIE\n")
	assertNoErrors(t)
	resetTestState()
}

func TestSwitch_CanSelectCaseBasedOnUsedVars(t *testing.T) {
	t.Skip("TODO RAD-141") // TODO RAD-141
	rsl := `
name = "alice"
age = 42
result = switch:
	case: "foo: {name}"
	case: "foo: {name}, bar: {age}"
	case: "foo: {name}, bar: {age}, baz: {notdefined}"
print(result)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "foo: alice, bar: 42\n")
	assertNoErrors(t)
	resetTestState()
}
