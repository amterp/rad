package testing

import "testing"

func TestNot(t *testing.T) {
	rsl := `
a = false
if !a:
    print("it works!")
if !!!a:
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
