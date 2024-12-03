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
	setupAndRunCode(t, "", "--VERSION")
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
	expected := `rad: A tool for writing user-friendly command line scripts.

Usage:
  rad [script path] [flags]

` + globalFlagHelp
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestMisc_Abs_Int(t *testing.T) {
	rsl := `
print(abs(10))
print(abs(-10))
`
	setupAndRunCode(t, rsl)
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
	setupAndRunCode(t, rsl)
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
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/7 on 'abs': abs() takes an integer or float, got string\n")
	resetTestState()
}
