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

L1:11: ERROR

     1 | hello = 2 a
       |           ^ Undefined identifier 'a'
       |           (code: RAD20028)
       = help: did you mean 'abs', 'map', or 'max'?

L3:6: ERROR

     3 | 	yes no
       |      ^ Undefined identifier 'no'
       |      (code: RAD20028)
       = help: did you mean 'now', 'int', or 'pow'?

Reported 4 diagnostics.
`
	setupAndRunArgs(t, "check", "./rad_scripts/invalid.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Check_UnknownFunctions(t *testing.T) {
	setupAndRunArgs(t, "check", "./rad_scripts/unknown_functions.rad", "--color=never")
	expected := `L1:1: ERROR

     1 | foo()
       | ^ Undefined identifier 'foo'
       | (code: RAD20028)
       = help: did you mean 'floor', 'now', or 'pow'?

L3:1: ERROR

     3 | qux()
       | ^ Undefined identifier 'qux'
       | (code: RAD20028)
       = help: did you mean 'max' or 'sum'?

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
	// A named callback is a plain function reference, so an unknown
	// target is the standard undefined-identifier error - same as any
	// other reference to a name that isn't in scope.
	setupAndRunArgs(t, "check", "./rad_scripts/unknown_command_callbacks.rad", "--color=never")
	expected := `L4:11: ERROR

     4 |     calls missing_one
       |           ^ Undefined identifier 'missing_one'
       |           (code: RAD20028)

L7:11: ERROR

     7 |     calls missing_two
       |           ^ Undefined identifier 'missing_two'
       |           (code: RAD20028)

Reported 2 diagnostics.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	// Unknown callbacks are now errors, so `rad check` exits non-zero.
	assertExitCode(t, 1)
}

func Test_Check_UnhandledFallible_HiddenByDefault(t *testing.T) {
	// RAD30011 (unhandled fallible call) is suppressed by default - it's too
	// noisy to be on for every fallible builtin.
	expected := `No diagnostics to report.
`
	setupAndRunArgs(t, "check", "./rad_scripts/unhandled_fallible.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Check_UnhandledFallible_StrictSurfacesWarning(t *testing.T) {
	// Under --strict the advisory surfaces, as a warning (not an error), so it
	// doesn't fail the exit code.
	expected := `L1:5: WARN

     1 | n = parse_int("5")
       |     ^ This call can fail; the error isn't handled and would halt the script
       |     (code: RAD30011)
       = help: Handle it with ` + "`catch`" + ` or ` + "`??`" + `, e.g. ` + "`x = <call> catch <fallback>`" + `

Reported 1 diagnostic.
`
	setupAndRunArgs(t, "check", "./rad_scripts/unhandled_fallible.rad", "--strict", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 0)
}
