package testing

import "testing"

func TestSyntaxError(t *testing.T) {
	setupAndRunArgs(t, "./test_rads/invalid_syntax.rad")
	expected := "RslError at L1/1 on '1': Expected Identifier\n"
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertExitCode(t, 1)
	resetTestState()
}
