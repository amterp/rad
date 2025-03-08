package testing

import "testing"

func TestString_SimpleIndexing(t *testing.T) {
	rsl := `
a = "alice"
print(a[0])
print(a[1])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\nl\n")
	assertNoErrors(t)
	resetTestState()
}

func TestString_NegativeIndexing(t *testing.T) {
	rsl := `
a = "alice"
print(a[-1])
print(a[-2])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "e\nc\n")
	assertNoErrors(t)
	resetTestState()
}

func TestString_ComplexIndexing(t *testing.T) {
	rsl := `
a = "alice"
print(a[len(a) - 3])
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "i\n")
	assertNoErrors(t)
	resetTestState()
}
