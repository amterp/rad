package testing

import "testing"

func TestStartsEndsContains(t *testing.T) {
	setupAndRunArgs(t, "./rads/starts_ends_contains.rad")
	expected := `true
false
false
true
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
