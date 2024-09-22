package testing

import "testing"

func TestSyntaxError(t *testing.T) {
	setupAndRunArgs(t, "./test_rads/invalid_syntax.rad")
	assertError(t, 1, "RslError at L1/1 on '1': Expected Identifier\n")
	resetTestState()
}
