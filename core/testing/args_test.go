package testing

import (
	"testing"
)

const (
	setupArgScript = `
args:
   foo "bar" x string`
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
  <bar> [OPTIONS]

Script args:
  -x, --bar string   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_PrintsUsageWithoutErrorIfNoArgsPassedOneRequiredOneOptionalArg(t *testing.T) {
	script := `
args:
	mandatory string
	optional int = 10
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Usage:
  <mandatory> [optional] [OPTIONS]

Script args:
      --mandatory string   
      --optional int        (default 10)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_InvokesIfNoArgsPassedButAllArgsAreOptional(t *testing.T) {
	t.Skip("Optional args temporarily not supported -- need to rethink")
	script := `
args:
	optionalS string?
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
	mandatory1 string
	mandatory2 string
	optional int = 10
print('hi')
`
	setupAndRunCode(t, script, "one", "--color=never")
	expected := `Missing required arguments: [mandatory2]

Usage:
  <mandatory1> <mandatory2> [optional] [OPTIONS]

Script args:
      --mandatory1 string   
      --mandatory2 string   
      --optional int         (default 10)

` + scriptGlobalFlagHelp
	assertOutput(t, stdOutBuffer, "")
	assertError(t, 1, expected)
}

func TestArgs_CanParseAllTypes(t *testing.T) {
	script := `
args:
    stringArg string
    intArg int
    floatArg float
    boolArg bool
    stringArrayArg string[]
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
	setupAndRunCode(t, script, "alice", "1", "1.1", "bob,charlie", "2,3", "2.1,3.1", "true,false", "--boolArg")
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
	stringArg string = "alice"
	intArg int = 1
	floatArg float = 1.1
	boolArg bool = true
	stringArrayArg string[] = ["bob", "charlie"]
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
	name string # The name.
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  <name> [OPTIONS]

Script args:
      --name string   The name.

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_ShortHelp(t *testing.T) {
	script := `
args:
	name string # The name.
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  <name> [OPTIONS]

Script args:
      --name string   The name.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_ShortHelpNoArgs(t *testing.T) {
	script := `
print("hi")
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  [OPTIONS]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_FullHelpNoArgs(t *testing.T) {
	script := `
print("hi")
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  [OPTIONS]

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_HelpWorksForAllTypes(t *testing.T) {
	script := `
args:
	stringArg string = "alice"
	intArg int = 1 # An int.
	floatArg float = 1.1
	boolArg bool = true
	stringArrayArg string[] = ["bob", "charlie"]
	intArrayArg int[] = [2, 3]
	floatArrayArg float[] = [2.1, 3.1]
	boolArrayArg bool[] = [true, false]
`
	setupAndRunCode(t, script, "-h", "--color=never")
	expected := `Usage:
  [stringArg] [intArg] [floatArg] [stringArrayArg] [intArrayArg] [floatArrayArg] [boolArrayArg] [OPTIONS]

Script args:
      --stringArg string                (default alice)
      --intArg int                     An int. (default 1)
      --floatArg float                  (default 1.1)
      --stringArrayArg string,string    (default ["bob", "charlie"])
      --intArrayArg int,int             (default [2, 3])
      --floatArrayArg float,float       (default [2.1, 3.1])
      --boolArrayArg bool,bool          (default [true, false])
      --boolArg                         (default true)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_UnsetBoolDefaultsToFalse(t *testing.T) {
	script := `
args:
	name string
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
	name string
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
	name string
	age int
`
	setupAndRunCode(t, script, "alice", "--color=never")
	expected := `Missing required arguments: [age]

Usage:
  <name> <age> [OPTIONS]

Script args:
      --name string   
      --age int       

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func TestArgs_TooManyArgsPrintsUsageAndReturnsError(t *testing.T) {
	script := `
args:
	name string
	age int
`
	setupAndRunCode(t, script, "alice", "2", "3", "--color=never")
	expected := `Too many positional arguments. Unused: [3]

Usage:
  <name> <age> [OPTIONS]

Script args:
      --name string   
      --age int       

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func TestArgs_InvalidFlagPrintsUsageAndReturnsError(t *testing.T) {
	script := `
args:
	name string
	age int
`
	setupAndRunCode(t, script, "alice", "2", "-s", "--color=never")
	expected := `unknown shorthand flag: 's' in -s

Usage:
  <name> <age> [OPTIONS]

Script args:
      --name string   
      --age int       

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}
