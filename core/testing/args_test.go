package testing

import (
	"testing"
)

const (
	setupArgScript = `
args:
   foo "bar" x str`
)

func TestArgs_ApiRename(t *testing.T) {
	script := setupArgScript + `
print(foo)
`
	setupAndRunCode(t, script, "hey")
	expected := `hey
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_ApiRenameUsageString(t *testing.T) {
	setupAndRunCode(t, setupArgScript, "-h", "--color=never")
	expected := `Usage:
  TestCase <bar> [OPTIONS]

Script args:
  -x, --bar str
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 0)
}

func TestArgs_PrintsUsageWithoutErrorIfNoArgsPassedOneRequiredOneOptionalArg(t *testing.T) {
	script := `
args:
	mandatory str
	optional int = 10
`
	setupAndRunCode(t, script)
	expected := "\x1b[32;1mUsage:\x1b[0;22m\n  \x1b[1mTestCase\x1b[22m \x1b[36m<mandatory>\x1b[0m \x1b[36m[optional]\x1b[0m \x1b[36m[OPTIONS]\x1b[0m\n\n\x1b[32;1mScript args:\x1b[0;22m"
	expected += `
      --mandatory str
      --optional int    (default 10)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 0)
}

func TestArgs_InvokesIfNoArgsPassedButAllArgsAreOptional(t *testing.T) {
	t.Skip("Optional args temporarily not supported -- need to rethink")
	script := `
args:
	optionalS str?
	optionalI int = 10
print('hi')
`
	setupAndRunCode(t, script, "--color=never")
	expected := `hi
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_ErrorsIfSomeRequiredArgsMissing(t *testing.T) {
	script := `
args:
	mandatory1 str
	mandatory2 str
	optional int = 10
print('hi')
`
	setupAndRunCode(t, script, "one", "--color=never")
	expected := `Missing required arguments: [mandatory2]

Usage:
  TestCase <mandatory1> <mandatory2> [optional] [OPTIONS]

Script args:
      --mandatory1 str
      --mandatory2 str
      --optional int     (default 10)

` + scriptGlobalFlagHelp
	assertOutput(t, stdOutBuffer, "")
	assertError(t, 1, expected)
}

func TestArgs_CanParseAllTypes(t *testing.T) {
	script := `
args:
    stringArg str
    intArg int
    floatArg float
    boolArg bool
    stringArrayArg str[]
    intArrayArg int[]
    floatArrayArg float[]
    boolArrayArg bool[]
print(upper(stringArg))
print(intArg + 1)
print(floatArg + 1.1)
print(boolArg or false)
print(upper(stringArrayArg[0]))
print(intArrayArg[0] + 1)
print(floatArrayArg[0] + 1.1)
print(boolArrayArg[0] or false)
`
	setupAndRunCode(t, script, "alice", "1", "1.1", "--stringArrayArg", "bob", "--stringArrayArg", "charlie", "--intArrayArg", "2", "--intArrayArg", "3", "--floatArrayArg", "2.1", "--floatArrayArg", "3.1", "--boolArrayArg", "true", "--boolArrayArg", "false", "--boolArg")
	expected := `ALICE
2
2.2
true
BOB
3
3.2
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanParseAllTypeDefaults(t *testing.T) {
	script := `
args:
	stringArg str = "alice"
	intArg int = 1
	floatArg float = 1.1
	boolArg bool = true
	stringArrayArg str[] = ["bob", "charlie"]
	intArrayArg int[] = [2, 3]
	floatArrayArg float[] = [2.1, 3.1]
	boolArrayArg bool[] = [true, false]
print(upper(stringArg))
print(intArg + 1)
print(floatArg + 1.1)
print(boolArg or false)
print(upper(stringArrayArg[0]))
print(intArrayArg[0] + 1)
print(floatArrayArg[0] + 1.1)
print(boolArrayArg[0] or false)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ALICE
2
2.2
true
BOB
3
3.2
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanHaveNegativeIntDefault(t *testing.T) {
	script := `
args:
	intArg int = -10
print(intArg + 1)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `-9
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanHaveNegativeFloatDefault(t *testing.T) {
	script := `
args:
	floatArg float = -10.2
print(floatArg + 1)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `-9.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeIntWithFlag(t *testing.T) {
	script := `
args:
	intArg int
print(intArg)
`
	setupAndRunCode(t, script, "--intArg", "-10")
	expected := `-10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeIntWithoutFlag(t *testing.T) {
	t.Skip("TODO: RAD-71") // todo RAD-71
	script := `
args:
	intArg int
print(intArg)
`
	setupAndRunCode(t, script, "-10")
	expected := `-10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeFloatWithFlag(t *testing.T) {
	script := `
args:
	floatArg float
print(floatArg)
`
	setupAndRunCode(t, script, "--floatArg", "-10.2")
	expected := `-10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeFloatWithoutFlag(t *testing.T) {
	t.Skip("TODO: RAD-71") // todo RAD-71
	script := `
args:
	floatArg float
print(floatArg)
`
	setupAndRunCode(t, script, "-10.2")
	expected := `-10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanHaveSeveralMinuses(t *testing.T) {
	script := `
args:
	intArg int = --- 10
	floatArg float = -------10.2
print(intArg + 1)
print(floatArg + 1)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `-9
-9.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanHaveIntAsDefaultForFloatArg(t *testing.T) {
	script := `
args:
	floatArg float = 2
print(floatArg + 1.2)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `3.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CannotHaveFloatAsDefaultForIntArg(t *testing.T) {
	script := `
args:
	intArg int = 1.2
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:16

  	intArg int = 1.2
                 ^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func TestArgs_FullHelp(t *testing.T) {
	script := `
args:
	name str # The name.
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str   The name.

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_ShortHelp(t *testing.T) {
	script := `
args:
	name str # The name.
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str   The name.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_ShortHelpNoArgs(t *testing.T) {
	script := `
print("hi")
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  TestCase [OPTIONS]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_FullHelpNoArgs(t *testing.T) {
	script := `
print("hi")
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  TestCase [OPTIONS]

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_HelpWorksForAllTypes(t *testing.T) {
	script := `
args:
	stringArg str = "alice"
	intArg int = 1 # An int.
	floatArg float = 1.1
	boolArg bool = true
	stringListArg str[] = ["bob", "charlie"]
	intListArg int[] = [2, 3]
	floatListArg float[] = [2.1, 3.1]
	boolListArg bool[] = [true, false]
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  TestCase [stringArg] [intArg] [floatArg] [stringListArg] [intListArg] [floatListArg] [boolListArg] [OPTIONS]

Script args:
      --stringArg str         (default alice)
      --intArg int            An int. (default 1)
      --floatArg float        (default 1.1)
      --stringListArg strs    (default [bob, charlie])
      --intListArg ints       (default [2, 3])
      --floatListArg floats   (default [2.1, 3.1])
      --boolListArg bools     (default [true, false])
      --boolArg               (default true)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_UnsetBoolDefaultsToFalse(t *testing.T) {
	script := `
args:
	name str
	isTall bool
print(name, isTall)
`
	setupAndRunCode(t, script, "alice")
	expected := `alice false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_CanDefaultBoolToTrue(t *testing.T) {
	script := `
args:
	name str
	isTall bool = true
print(name, isTall)
`
	setupAndRunCode(t, script, "alice")
	expected := `alice true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_MissingArgsPrintsUsageAndReturnsError(t *testing.T) {
	script := `
args:
	name str
	age int
`
	setupAndRunCode(t, script, "alice", "--color=never")
	expected := `Missing required arguments: [age]

Usage:
  TestCase <name> <age> [OPTIONS]

Script args:
      --name str
      --age int

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func TestArgs_TooManyArgsPrintsUsageAndReturnsError(t *testing.T) {
	script := `
args:
	name str
	age int
`
	setupAndRunCode(t, script, "alice", "2", "3", "--color=never")
	expected := `Too many positional arguments. Unused: [3]

Usage:
  TestCase <name> <age> [OPTIONS]

Script args:
      --name str
      --age int

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func TestArgs_InvalidFlagPrintsUsageAndReturnsError(t *testing.T) {
	script := `
args:
	name str
	age int
`
	setupAndRunCode(t, script, "alice", "2", "-s", "--color=never")
	expected := `unknown shorthand flag: 's' in -s

Usage:
  TestCase <name> <age> [OPTIONS]

Script args:
      --name str
      --age int

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func Test_Args_AutomaticallyReplacesUnderscoresWithHyphens(t *testing.T) {
	script := `
args:
	test_arg str
print(test_arg)
`
	setupAndRunCode(t, script, "--test-arg", "bob", "--color=never")
	expected := `bob
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func Test_Args_AutomaticallyReplacesUnderscoresWithHyphensUsage(t *testing.T) {
	script := `
args:
	test_arg str
print(test_arg)
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  TestCase <test-arg> [OPTIONS]

Script args:
      --test-arg str
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}
