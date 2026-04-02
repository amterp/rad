package testing

import "testing"

func Test_Check_Valid(t *testing.T) {
	// todo should be more happy about it!
	expected := `No diagnostics to report.
`
	setupAndRunArgs(t, "check", "./rad_scripts/hello.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Check(t *testing.T) {
	expected := `L1:9: ERROR

     1 | hello = 2 a
       |         ^ Unexpected '2'
       |         (code: RAD10009)

L3:2: ERROR

     3 | 	yes no
       |  ^ Unexpected 'yes'
       |  (code: RAD10009)

Reported 2 diagnostics.
`
	setupAndRunArgs(t, "check", "./rad_scripts/invalid.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Check_UnknownFunctions(t *testing.T) {
	setupAndRunArgs(t, "check", "./rad_scripts/unknown_functions.rad", "--color=never")
	expected := `L1:1: HINT

     1 | foo()
       | ^ Function 'foo' may not be defined (only built-in and top-level functions are tracked)
       | (code: RAD40003)

L3:1: HINT

     3 | qux()
       | ^ Function 'qux' may not be defined (only built-in and top-level functions are tracked)
       | (code: RAD40003)

Reported 2 diagnostics.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func Test_Check_RadOptionNoEffect(t *testing.T) {
	setupAndRunArgs(t, "check", "./rad_scripts/rad_option_no_effect.rad", "--color=never")
	expected := `L3:5: WARN

     3 |     insecure
       |     ^ 'insecure' has no effect without a URL source
       |     (code: RAD40007)

L4:5: WARN

     4 |     quiet
       |     ^ 'quiet' has no effect without a URL source
       |     (code: RAD40007)

L5:5: WARN

     5 |     noprint
       |     ^ 'noprint' has no effect without a source (mutations are not preserved)
       |     (code: RAD40007)

Reported 3 diagnostics.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func Test_Check_DeprecatedRequest(t *testing.T) {
	setupAndRunArgs(t, "check", "./rad_scripts/deprecated_request.rad", "--color=never")
	expected := `L1:1: ERROR

     1 | request "http://example.com":
       | ^ 'request' blocks have been removed. Use 'rad' instead. See https://amterp.dev/rad/migrations/v0.9/
       | (code: RAD40008)

Reported 1 diagnostic.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Check_DeprecatedDisplay(t *testing.T) {
	setupAndRunArgs(t, "check", "./rad_scripts/deprecated_display.rad", "--color=never")
	expected := `L2:1: ERROR

     2 | display data:
       | ^ 'display' blocks have been removed. Use 'rad' instead. See https://amterp.dev/rad/migrations/v0.9/
       | (code: RAD40008)

Reported 1 diagnostic.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Check_UnknownCommandCallbacks(t *testing.T) {
	setupAndRunArgs(t, "check", "./rad_scripts/unknown_command_callbacks.rad", "--color=never")
	expected := `L4:11: WARN

     4 |     calls missing_one
       |           ^ Function 'missing_one' may not be defined (only built-in and top-level functions are tracked)
       |           (code: RAD40003)

L7:11: WARN

     7 |     calls missing_two
       |           ^ Function 'missing_two' may not be defined (only built-in and top-level functions are tracked)
       |           (code: RAD40003)

Reported 2 diagnostics.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}
