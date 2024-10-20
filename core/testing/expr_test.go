package testing

import "testing"

func TestNot(t *testing.T) {
	rsl := `
a = false
if not a:
    print("it works!")
if not not not a:
    print("it works!!!")
`
	setupAndRunCode(t, rsl)
	expected := `it works!
it works!!!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
