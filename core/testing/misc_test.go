package testing

import (
	"rad/core"
	"testing"
)

func TestMisc_SyntaxError(t *testing.T) {
	setupAndRunArgs(t, "./rsl_scripts/invalid_syntax.rad")
	assertError(t, 1, "RslError at L1/1 on '1': Expected Identifier\n")
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
	setupAndRunCode(t, "", "--version")
	assertOnlyOutput(t, stdErrBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_VersionShort(t *testing.T) {
	setupAndRunCode(t, "", "-V")
	assertOnlyOutput(t, stdErrBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_PrioritizesHelpIfBothHelpAndVersionSpecified(t *testing.T) {
	setupAndRunCode(t, "", "-h", "-V", "--NO-COLOR")
	expected := `Request And Display (RAD): A tool for writing user-friendly command line scripts.

Usage:
  rad [script path] [flags]

Global flags:
  -h, --help                   Print usage string.
  -D, --DEBUG                  Enables debug output. Intended for RSL script developers.
      --RAD-DEBUG              Enables Rad debug output. Intended for Rad developers.
      --NO-COLOR               Disable colorized output.
  -Q, --QUIET                  Suppresses some output.
      --SHELL                  Outputs shell/bash exports of variables, so they can be eval'd
  -V, --version                Print rad version information.
      --STDIN script-name      Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.
      --MOCK-RESPONSE string   Add mock response for json requests (pattern:filePath)
`
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
