package testing

import (
	"testing"
)

func TestPrint(t *testing.T) {
	setupAndRunArgs(t, "./rads/print.rad")
	expected := `hi alice
hi bob
hi charlie
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestDebugNoDebugFlag(t *testing.T) {
	setupAndRunArgs(t, "./rads/debug.rad")
	expected := "one\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestDebugWithDebugFlag(t *testing.T) {
	setupAndRunArgs(t, "./rads/debug.rad", "--DEBUG")
	expected := "one\nDEBUG: two\nDEBUG: three\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
