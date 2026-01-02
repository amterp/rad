package testing

import (
	"strings"
	"testing"

	"github.com/amterp/rad/core"
)

func Test_Misc_SyntaxError(t *testing.T) {
	setupAndRunCode(t, "1 = 2", "--color=never")
	expected := `Error at L1:1

  1 = 2
  ^^^ Unexpected '1 ='
`
	assertError(t, 1, expected)
}

func Test_Misc_ReadingFromUnderscoreVarErrors(t *testing.T) {
	script := `
_ = 2
print(_)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:7

  print(_)
        ^ Cannot use '_' as a value
`
	assertError(t, 1, expected)
}

func Test_Misc_CanHaveVarNameThatIsJustAnUnderscoreInForLoop(t *testing.T) {
	script := `
a = [1, 2, 3]
for _, v in a:
	print(v)
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
	assertOnlyOutput(t, stdOutBuffer, "rad "+core.Version+"\n")
	assertNoErrors(t)
}

func Test_Misc_VersionShort(t *testing.T) {
	setupAndRunCode(t, "", "-v")
	assertOnlyOutput(t, stdOutBuffer, "rad "+core.Version+"\n")
	assertNoErrors(t)
}

func Test_Misc_PrioritizesHelpIfBothHelpAndVersionSpecified(t *testing.T) {
	t.Skip("TODO: currently failing, should fix")
	setupAndRunCode(t, "", "-h", "-v", "--color=never")
	expected := radHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_PrintsHelpToStderrIfUnknownGlobalFlag(t *testing.T) {
	setupAndRunArgs(t, "--asd", "--color=never")
	expected := "Unknown arguments: [--asd]\n\n\n" + radHelp
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
          ^^^^^
          Value '"asd"' (str) is not compatible with expected type 'int|float'
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
  TestCase <src> [OPTIONS]

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
  -r, --repl            Start interactive REPL mode.
  -d                    Enables debug output. Intended for Rad script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never] (default auto)
  -q, --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`
	expected := `Usage:
  TestCase <debug> [OPTIONS]

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
  -r, --repl            Start interactive REPL mode.
  -d, --debug           Enables debug output. Intended for Rad script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never] (default auto)
      --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`
	expected := `Usage:
  TestCase <myquiet> [OPTIONS]

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
  TestCase <version> [OPTIONS]

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

func Test_Misc_UserSrcFlagShadowsGlobalSrcFlag(t *testing.T) {
	script := `
args:
	src str

print("User src value:", src)
`
	setupAndRunCode(t, script, "--src", "user-provided-value", "--color=never")
	expected := `User src value: user-provided-value
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_GlobalVersionFlagBypassesValidation(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--version", "--color=never")
	expected := "rad " + core.Version + "\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_GlobalSrcTreeFlagBypassesValidation(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--src-tree", "--color=never")
	// Just check that it starts with the expected tree format and doesn't error
	output := stdOutBuffer.String()
	if !strings.Contains(output, "source_file") || !strings.Contains(output, "arg_block") {
		t.Errorf("Expected syntax tree output, got: %s", output)
	}
	assertNoErrors(t)
}

func Test_Misc_GlobalRadArgsDumpFlag(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/example_arg.rad", "--rad-args-dump", "--color=never")
	output := stdOutBuffer.String()
	if !strings.Contains(output, "Ra Command Dump") {
		t.Errorf("Expected Ra dump in output, got: %s", output)
	}
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

func Test_Misc_PercentCharactersInFileHeader(t *testing.T) {
	script := `---
This script calculates 10% of values.
It shows 100% accurate results.
URL encoding like %20 also works fine.
---
args:
    value int    # Input value for % calculation

print("Result: {value}")
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `This script calculates 10% of values.
It shows 100% accurate results.
URL encoding like %20 also works fine.

Usage:
  TestCase <value> [OPTIONS]

Script args:
      --value int   Input value for % calculation

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_PercentCharactersInArgComments(t *testing.T) {
	script := `---
Test script for percent in arg comments
---
args:
    percentage float    # Value as %, e.g. 95.5%
    url str            # URL with %20 encoding
    discount int       # Discount rate (5% to 50%)

print("Values: {percentage}, {url}, {discount}")
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Test script for percent in arg comments

Usage:
  TestCase <percentage> <url> <discount> [OPTIONS]

Script args:
      --percentage float   Value as %, e.g. 95.5%
      --url str            URL with %20 encoding
      --discount int       Discount rate (5% to 50%)

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_DoesNotMisFormatWithMissing(t *testing.T) {
	script := `pprint("%s")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `"%s"
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_InvalidSyntax_WithSrcFlag(t *testing.T) {
	script := `foo = [11, 12, 13
`
	setupAndRunCode(t, script, "--src", "--color=never")
	expected := `foo = [11, 12, 13

`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Misc_InvalidSyntax_WithSrcTreeFlag(t *testing.T) {
	script := `foo = [11, 12, 13
`
	setupAndRunCode(t, script, "--src-tree", "--color=never")
	output := stdOutBuffer.String()
	if !strings.Contains(output, "source_file") || !strings.Contains(output, "ERROR") {
		t.Errorf("Expected syntax tree with ERROR node, got: %s", output)
	}
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
