package testing

import (
	"rad/core"
	"testing"
)

func TestMisc_SyntaxError(t *testing.T) {
	setupAndRunCode(t, "1 = 2", "--NO-COLOR")
	expected := `Error at L1:1

  1 = 2
  ^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestMisc_CanHaveVarNameThatIsJustAnUnderscore(t *testing.T) {
	rsl := `
_ = 2
print(_)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_CanHaveVarNameThatIsJustAnUnderscoreInForLoop(t *testing.T) {
	rsl := `
a = [1, 2, 3]
for _, _ in a:
	print(_)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_CanHaveNegativeNumbers(t *testing.T) {
	rsl := `
a = -10
print(a)
b = -20.2
print(b)
print("{-12}")
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "-10\n-20.2\n-12\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_Version(t *testing.T) {
	setupAndRunCode(t, "", "--VERSION")
	assertOnlyOutput(t, stdOutBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_VersionShort(t *testing.T) {
	setupAndRunCode(t, "", "-V")
	assertOnlyOutput(t, stdOutBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_PrioritizesHelpIfBothHelpAndVersionSpecified(t *testing.T) {
	setupAndRunCode(t, "", "-h", "-V", "--NO-COLOR")
	expected := `rad: A tool for writing user-friendly command line scripts.
GitHub: https://github.com/amterp/rad
Documentation: https://amterp.github.io/rad/

Usage:
  rad [script path | command] [flags]

Commands:
  new           Sets up a new RSL script, including some boilerplate and execution permissions.

To see help for a specific command, run ` + "`rad <command> -h`.\n\n" + globalFlagHelp + `
To execute an RSL script:
  rad path/to/script.rsl [args]

To execute a command:
  rad <command> [args]

If you're new, check out the Getting Started guide: https://amterp.github.io/rad/guide/getting-started/
`
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_Abs_Int(t *testing.T) {
	rsl := `
print(abs(10))
print(abs(-10))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `10
10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_Abs_Float(t *testing.T) {
	rsl := `
print(abs(10.2))
print(abs(-10.2))
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `10.2
10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_Abs_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = abs("asd")
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L2:9

  a = abs("asd")
          ^^^^^
          Got "string" as the 1st argument of abs(), but must be: float or int
`
	assertError(t, 1, expected)
	resetTestState()
}
