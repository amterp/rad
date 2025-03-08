package testing

import (
	"rad/core"
	"strings"
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

func Test_Misc_CanShadowGlobalFlag(t *testing.T) {
	rsl := `
args:
	SRC string
`
	setupAndRunCode(t, rsl, "--COLOR=never", "-h")
	expectedGlobalFlags := globalFlagHelpWithout("SRC")
	expected := `Usage:
  <SRC>

Script args:
      --SRC string   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_CanShadowGlobalShorthand(t *testing.T) {
	rsl := `
args:
	version V string
`
	setupAndRunCode(t, rsl, "--COLOR=never", "-h")
	expectedGlobalFlags := `Global flags:
  -h, --help                   Print usage string.
  -D, --DEBUG                  Enables debug output. Intended for RSL script developers.
      --RAD-DEBUG              Enables Rad debug output. Intended for Rad developers.
      --COLOR mode             Control output colorization. Valid values: [auto, always, never]. (default auto)
  -Q, --QUIET                  Suppresses some output.
      --SHELL                  Outputs shell/bash exports of variables, so they can be eval'd
      --VERSION                Print rad version information.
      --CONFIRM-SHELL          Confirm all shell commands before running them.
      --SRC                    Instead of running the target script, just print it out.
      --RSL-TREE               Instead of running the target script, print out its syntax tree.
      --MOCK-RESPONSE string   Add mock response for json requests (pattern:filePath)
`
	expected := `Usage:
  <version>

Script args:
  -V, --version string   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_CanShadowGlobalFlagAndShorthand(t *testing.T) {
	rsl := `
args:
	VERSION V string
`
	setupAndRunCode(t, rsl, "--COLOR=never", "-h")
	expectedGlobalFlags := globalFlagHelpWithout("VERSION")
	expected := `Usage:
  <VERSION>

Script args:
  -V, --VERSION string   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Misc_CanShadowGlobalFlagAndUseIt(t *testing.T) {
	rsl := `
args:
	VERSION V string
print(VERSION+"!")
`
	setupAndRunCode(t, rsl, "someversion", "--COLOR=never")
	expected := `someversion!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func globalFlagHelpWithout(s string) string {
	original := globalFlagHelp
	removeLineWith := "--" + s
	lines := strings.Split(original, "\n")
	var result []string
	for _, line := range lines {
		if !strings.Contains(line, removeLineWith) {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}
