package testing

import (
	"testing"
)

func TestPrint(t *testing.T) {
	setupAndRun(t, "./test_rads/print.rad")

	expected := `hi alice
hi bob
hi charlie
`
	assertOnly(t, stdOutBuffer, expected)
}

func TestDebugNoDebugFlag(t *testing.T) {
	setupAndRun(t, "./test_rads/debug.rad")
	expected := "one\n"
	assertOnly(t, stdOutBuffer, expected)
}

func TestDebugWithDebugFlag(t *testing.T) {
	setupAndRun(t, "./test_rads/debug.rad", "--DEBUG")
	expected := "one\nDEBUG: two\nDEBUG: three\n"
	assertOnly(t, stdOutBuffer, expected)
}
