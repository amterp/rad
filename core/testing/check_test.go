package testing

import "testing"

func Test_Check_Valid(t *testing.T) {
	// todo should be more happy about it!
	expected := `No diagnostics to report.
`
	setupAndRunArgs(t, "check", "./rsl_scripts/hello.rsl", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Check(t *testing.T) {
	expected := `L1:9: ERROR

     1 | hello = 2 a
       |         ^ Invalid syntax
       |         (code: RAD10001)

L3:2: ERROR

     3 | 	yes no
       |  ^ Invalid syntax
       |  (code: RAD10001)

Reported 2 diagnostics.
`
	setupAndRunArgs(t, "check", "./rsl_scripts/invalid.rsl", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}
