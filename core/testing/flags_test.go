package testing

import "testing"

func TestArgs_CanPassPositiveInts(t *testing.T) {
	rsl := `
args:
	intArg int
print(intArg)
`
	setupAndRunCode(t, rsl, "2")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeInts(t *testing.T) {
	rsl := `
args:
	intArg int
print(intArg)
`
	// -- forces it to be positional so pflag does not think it's a flag. address?
	setupAndRunCode(t, rsl, "--", "-2")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
}
