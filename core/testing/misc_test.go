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
	script := `
_ = 2
print(_)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
}

func Test_Misc_CanHaveVarNameThatIsJustAnUnderscoreInForLoop(t *testing.T) {
	script := `
a = [1, 2, 3]
for _, _ in a:
	print(_)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n2\n3\n")
	assertNoErrors(t)
}

func Test_Misc_CanHaveNegativeNumbers(t *testing.T) {
	script := `
a = -10
print(a)
b = -20.2
print(b)
print("{-12}")
`
	setupAndRunCode(t, script, "--color=never")
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
	script := `
print(abs(10))
print(abs(-10))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10
10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_Abs_Float(t *testing.T) {
	script := `
print(abs(10.2))
print(abs(-10.2))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10.2
10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_Abs_ErrorsOnAlphabetical(t *testing.T) {
	script := `
a = abs("asd")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:9

  a = abs("asd")
          ^^^^^ Value '"asd"' (str) is not compatible with expected type 'float'
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
	script := `
args:
	src str
`
	setupAndRunCode(t, script, "--color=never", "--help")
	expectedGlobalFlags := globalFlagHelpWithout("src")
	expected := `Usage:
  <src> [OPTIONS]

Script args:
      --src str   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalFlagThatHasShorthand(t *testing.T) {
	script := `
args:
	debug str
`
	setupAndRunCode(t, script, "--color=never", "--help")
	expectedGlobalFlags := `Global options:
  -h, --help            Print usage string.
  -d                    Enables debug output. Intended for Rad script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never]. (default auto)
  -q, --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`
	expected := `Usage:
  <debug> [OPTIONS]

Script args:
      --debug str   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalShorthand(t *testing.T) {
	script := `
args:
	myquiet q str
`
	setupAndRunCode(t, script, "--color=never", "--help")
	expectedGlobalFlags := `Global options:
  -h, --help            Print usage string.
  -d, --debug           Enables debug output. Intended for Rad script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never]. (default auto)
      --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`
	expected := `Usage:
  <myquiet> [OPTIONS]

Script args:
  -q, --myquiet str   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalFlagAndShorthand(t *testing.T) {
	script := `
args:
	version v str
`
	setupAndRunCode(t, script, "--color=never", "--help")
	expectedGlobalFlags := globalFlagHelpWithout("version")
	expected := `Usage:
  <version> [OPTIONS]

Script args:
  -v, --version str   

` + expectedGlobalFlags
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_CanShadowGlobalFlagAndUseIt(t *testing.T) {
	script := `
args:
	version v str
print(version+"!")
`
	setupAndRunCode(t, script, "someversion", "--color=never")
	expected := `someversion!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_GlobalSrcFlag(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--src", "--color=never")
	expected := `args:
    name str # The name.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_Func_Get_Rad_Home(t *testing.T) {
	script := `
d = get_rad_home()
d = d.split("/")
print(d.len() > 0)
d[-1].print()
`
	setupAndRunCode(t, script)
	expected := `true
rad_test_home
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
