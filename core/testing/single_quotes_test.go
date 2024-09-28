package testing

import "testing"

func TestSingleQuotes(t *testing.T) {
	rsl := `
greeting = 'hi'
print(greeting)
name = "alice"
print(greeting + ' ' + name)
`
	setupAndRunCode(t, rsl)
	expected := `hi
hi alice
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
