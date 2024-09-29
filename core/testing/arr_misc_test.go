package testing

import "testing"

func TestCanUseVarsInArrays(t *testing.T) {
	rsl := `
a = "a"
b = 1
c = true
print([a, b, c])
`
	setupAndRunCode(t, rsl)
	expected := `[a, 1, true]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
