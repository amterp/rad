package testing

import "testing"

func TestForLoop(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
for item in a:
	print(item)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "a\nb\nc\n")
	assertNoErrors(t)
	resetTestState()
}

func TestForILoop(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
for idx, item in a:
	print(idx, item)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "0 a\n1 b\n2 c\n")
	assertNoErrors(t)
	resetTestState()
}

func TestOutsideChangesAreRemembered(t *testing.T) {
	rsl := `
num = 0
a = ["a", "b", "c"]
for idx, item in a:
	num += idx
print(num)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}
