package testing

import "testing"

func TestStartsEndsContains(t *testing.T) {
	rsl := `
a = "alice"
print(starts_with(a, "al"))
print(starts_with(a, "ce"))

print(ends_with(a, "al"))
print(ends_with(a, "ce"))
`
	setupAndRunCode(t, rsl)
	expected := `true
false
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
