package testing

import "testing"

func Test_StartsWith(t *testing.T) {
	rsl := `
a = "alice"
print(starts_with(a, "al"))
print(starts_with(a, "ce"))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_EndsWithWith(t *testing.T) {
	rsl := `
a = "alice"
print(ends_with(a, "al"))
print(ends_with(a, "ce"))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
