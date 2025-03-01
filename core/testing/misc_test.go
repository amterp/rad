package testing

import (
	"rad/core"
	"testing"
)

func Test_Misc_SyntaxError(t *testing.T) {
	setupAndRunCode(t, "1 = 2", "--COLOR=never")
	expected := `Error at L1:1

  1 = 2
  ^^^ Invalid syntax
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Misc_CanHaveVarNameThatIsJustAnUnderscore(t *testing.T) {
	rsl := `
_ = 2
print(_)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_CanHaveVarNameThatIsJustAnUnderscoreInForLoop(t *testing.T) {
	rsl := `
a = [1, 2, 3]
for _, _ in a:
	print(_)
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_CanHaveNegativeNumbers(t *testing.T) {
	rsl := `
a = -10
print(a)
b = -20.2
print(b)
print("{-12}")
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "-10\n-20.2\n-12\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_Version(t *testing.T) {
	setupAndRunCode(t, "", "--VERSION")
	assertOnlyOutput(t, stdOutBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_VersionShort(t *testing.T) {
	setupAndRunCode(t, "", "-V")
	assertOnlyOutput(t, stdOutBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_PrioritizesHelpIfBothHelpAndVersionSpecified(t *testing.T) {
	setupAndRunCode(t, "", "-h", "-V", "--COLOR=never")
	expected := radHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_PrintsHelpToStderrIfUnknownGlobalFlag(t *testing.T) {
	setupAndRunArgs(t, "--asd", "--COLOR=never")
	expected := "unknown flag: --asd\n" + radHelp
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Misc_Abs_Int(t *testing.T) {
	rsl := `
print(abs(10))
print(abs(-10))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `10
10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_Abs_Float(t *testing.T) {
	rsl := `
print(abs(10.2))
print(abs(-10.2))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `10.2
10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_Abs_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = abs("asd")
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L2:9

  a = abs("asd")
          ^^^^^
          Got "string" as the 1st argument of abs(), but must be: float or int
`
	assertError(t, 1, expected)
	resetTestState()
}

func Test_Misc_PrintsUsageIfInvokedWithNoScript(t *testing.T) {
	setupAndRunArgs(t, "--COLOR=never")
	expected := radHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
