package testing

import "testing"

func TestString_SimpleIndexing(t *testing.T) {
	script := `
a = "alice"
print(a[0])
print(a[1])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\nl\n")
	assertNoErrors(t)
}

func TestString_NegativeIndexing(t *testing.T) {
	script := `
a = "alice"
print(a[-1])
print(a[-2])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "e\nc\n")
	assertNoErrors(t)
}

func TestString_ComplexIndexing(t *testing.T) {
	script := `
a = "alice"
print(a[len(a) - 3])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "i\n")
	assertNoErrors(t)
}
