package testing

import "testing"

func TestCobra_CanPassPositiveInts(t *testing.T) {
	rsl := `
args:
	intArg int
print(intArg)
`
	setupAndRunCode(t, rsl, "2")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCobra_CanPassNegativeInts(t *testing.T) {
	rsl := `
args:
	intArg int
print(intArg)
`
	// -- forces it to be positional so Cobra does not think it's a flag. address?
	setupAndRunCode(t, rsl, "--", "-2")
	assertOnlyOutput(t, stdOutBuffer, "-2\n")
	assertNoErrors(t)
	resetTestState()
}
