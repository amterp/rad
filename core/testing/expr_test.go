package testing

import "testing"

func Test_Not(t *testing.T) {
	rsl := `
a = false
if not a:
    print("it works!")
if not not not a:
    print("it works!!!")
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `it works!
it works!!!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_ExprThenIndexing(t *testing.T) {
	rsl := `
a = [4, 2, 3, 1]
print(sort(a)[0])
print(sort(a)[2:][-1])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `1
4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
