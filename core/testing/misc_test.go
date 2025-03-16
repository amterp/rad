package testing

import (
	"rad/core"
	"strings"
	"testing"
)

func Test_Misc_SyntaxError(t *testing.T) {
	setupAndRunCode(t, "1 = 2", "--color=never")
	expected := `Error at L1:1

  1 = 2
  ^^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func Test_Misc_CanHaveVarNameThatIsJustAnUnderscore(t *testing.T) {
	rsl := `
_ = 2
print(_)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Misc_CanHaveVarNameThatIsJustAnUnderscoreInForLoop(t *testing.T) {
	rsl := `
a = [1, 2, 3]
for _, _ in a:
	print(_)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n")
	assertNoErrors(t)
}

func Test_Misc_CanHaveNegativeNumbers(t *testing.T) {
	rsl := `
a = -10
print(a)
b = -20.2
print(b)
print("{-12}")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "-10\n-20.2\n-12\n")
	assertNoErrors(t)
}

func Test_Misc_Version(t *testing.T) {
	setupAndRunCode(t, "", "--version")
	assertOnlyOutput(t, stdOutBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
}

func Test_Misc_VersionShort(t *testing.T) {
	setupAndRunCode(t, "", "-v")
	assertOnlyOutput(t, stdOutBuffer, "rad version "+core.Version+"\n")
	assertNoErrors(t)
}

func Test_Misc_PrioritizesHelpIfBothHelpAndVersionSpecified(t *testing.T) {
	setupAndRunCode(t, "", "-h", "-v", "--color=never")
	expected := radHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_PrintsHelpToStderrIfUnknownGlobalFlag(t *testing.T) {
	setupAndRunArgs(t, "--asd", "--color=never")
	expected := "unknown flag: --asd\n\n" + radHelp
	assertError(t, 1, expected)
}

func Test_Misc_Abs_Int(t *testing.T) {
	rsl := `
print(abs(10))
print(abs(-10))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `10
10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_Abs_Float(t *testing.T) {
	rsl := `
print(abs(10.2))
print(abs(-10.2))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `10.2
10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_Abs_ErrorsOnAlphabetical(t *testing.T) {
	rsl := `
a = abs("asd")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:9

  a = abs("asd")
          ^^^^^
          Got "string" as the 1st argument of abs(), but must be: float or int
`
	assertError(t, 1, expected)
}

func Test_Misc_PrintsUsageIfInvokedWithNoScript(t *testing.T) {
	setupAndRunArgs(t, "--color=never")
	expected := radHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalFlag(t *testing.T) {
	rsl := `
args:
	src string
`
	setupAndRunCode(t, rsl, "--color=never", "-h")
	expectedGlobalFlags := globalFlagHelpWithout("src")
	expected := `Usage:
  <src>

Script args:
      --src string   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalFlagThatHasShorthand(t *testing.T) {
	rsl := `
args:
	debug string
`
	setupAndRunCode(t, rsl, "--color=never", "-h")
	expectedGlobalFlags := `Global flags:
  -h, --help            Print usage string.
  -d                    Enables debug output. Intended for RSL script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never]. (default auto)
  -q, --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`
	expected := `Usage:
  <debug>

Script args:
      --debug string   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalShorthand(t *testing.T) {
	rsl := `
args:
	myquiet q string
`
	setupAndRunCode(t, rsl, "--color=never", "-h")
	expectedGlobalFlags := `Global flags:
  -h, --help            Print usage string.
  -d, --debug           Enables debug output. Intended for RSL script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never]. (default auto)
      --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`
	expected := `Usage:
  <myquiet>

Script args:
  -q, --myquiet string   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalFlagAndShorthand(t *testing.T) {
	rsl := `
args:
	version v string
`
	setupAndRunCode(t, rsl, "--color=never", "-h")
	expectedGlobalFlags := globalFlagHelpWithout("version")
	expected := `Usage:
  <version>

Script args:
  -v, --version string   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalFlagAndUseIt(t *testing.T) {
	rsl := `
args:
	version v string
print(version+"!")
`
	setupAndRunCode(t, rsl, "someversion", "--color=never")
	expected := `someversion!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_GlobalSrcFlag(t *testing.T) {
	setupAndRunArgs(t, "./rsl_scripts/example_arg.rsl", "--src", "--color=never")
	expected := `args:
    name string # The name.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func globalFlagHelpWithout(s string) string {
	original := scriptGlobalFlagHelp
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
