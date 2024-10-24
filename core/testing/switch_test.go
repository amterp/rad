package testing

import "testing"

func TestSwitch_CanSelectCaseBasedOnUsedVars(t *testing.T) {
	rsl := `
name = "alice"
age = 42
result = switch:
	case: "foo: {name}"
	case: "foo: {name}, bar: {age}"
	case: "foo: {name}, bar: {age}, baz: {notdefined}"
print(result)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "foo: alice, bar: 42\n")
	assertNoErrors(t)
	resetTestState()
}
