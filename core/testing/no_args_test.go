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
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestDebugNoDebugFlag(t *testing.T) {
	setupAndRun(t, "./test_rads/debug.rad")
	expected := "one\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestDebugWithDebugFlag(t *testing.T) {
	setupAndRun(t, "./test_rads/debug.rad", "--DEBUG")
	expected := "one\nDEBUG: two\nDEBUG: three\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStartsEndsContains(t *testing.T) {
	setupAndRun(t, "./test_rads/starts_ends_contains.rad")
	expected := `true
false
false
true
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestSyntaxError(t *testing.T) {
	setupAndRun(t, "./test_rads/invalid_syntax.rad")
	expected := "RslError at L1/1 on '1': Expected Identifier\n"
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertExitCode(t, 1)
}
