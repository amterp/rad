package testing

import "testing"

func TestArgs_CanPassPositiveInts(t *testing.T) {
	script := `
args:
	intArg int
print(intArg)
`
	setupAndRunCode(t, script, "2")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeInts(t *testing.T) {
	script := `
args:
	intArg int
print(intArg)
`
	// -- forces it to be positional so pflag does not think it's a flag. address?
	setupAndRunCode(t, script, "--", "-2")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
}
