package testing

import "testing"

func TestUnique(t *testing.T) {
	script := `
print(unique([2, 1, 2, 3, 1, "Alice", 4, 3, 5, 5]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 2, 1, 3, \"Alice\", 4, 5 ]\n")
	assertNoErrors(t)
}

func TestUnique_Large(t *testing.T) {
	script := `
a = unique([2 for i in range(1000)])
print(len(a))
print(a[0])
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n")
	assertNoErrors(t)
}

func TestUnique_String(t *testing.T) {
	script := `
print(join(unique(split("Frodo Baggins is a hobbit", "")), ""))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Frod Baginshbt\n")
	assertNoErrors(t)
}
