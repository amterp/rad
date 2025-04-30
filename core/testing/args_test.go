package testing

import (
	"testing"
)

const (
	setupArgRsl = `
args:
   foo "bar" x string`
)

func TestArgs_ApiRename(t *testing.T) {
	rsl := setupArgRsl + `
print(foo)
`
	setupAndRunCode(t, rsl, "hey")
	expected := `hey
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_ApiRenameUsageString(t *testing.T) {
	setupAndRunCode(t, setupArgRsl, "-h", "--color=never")
	expected := `Usage:
  <bar>

Script args:
  -x, --bar string   
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_PrintsUsageWithoutErrorIfNoArgsPassedOneRequiredOneOptionalArg(t *testing.T) {
	rsl := `
args:
	mandatory string
	optional int = 10
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Usage:
  <mandatory> [optional]

Script args:
      --mandatory string   
      --optional int        (default 10)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_InvokesIfNoArgsPassedButAllArgsAreOptional(t *testing.T) {
	t.Skip("Optional args temporarily not supported -- need to rethink")
	rsl := `
args:
	optionalS string?
	optionalI int = 10
print('hi')
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_ErrorsIfSomeRequiredArgsMissing(t *testing.T) {
	rsl := `
args:
	mandatory1 string
	mandatory2 string
	optional int = 10
print('hi')
`
	setupAndRunCode(t, rsl, "one", "--color=never")
	expected := `Missing required arguments: [mandatory2]

Usage:
  <mandatory1> <mandatory2> [optional]

Script args:
      --mandatory1 string   
      --mandatory2 string   
      --optional int         (default 10)

` + scriptGlobalFlagHelp
	assertOutput(t, stdOutBuffer, "")
	assertError(t, 1, expected)
}

func TestArgs_CanParseAllTypes(t *testing.T) {
	rsl := `
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
	setupAndRunCode(t, rsl, "alice", "1", "1.1", "true", "bob,charlie", "2,3", "2.1,3.1", "true,false")
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
	rsl := `
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
	setupAndRunCode(t, rsl, "--color=never")
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
	rsl := `
args:
	intArg int = -10
print(intArg + 1)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `-9
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanHaveNegativeFloatDefault(t *testing.T) {
	rsl := `
args:
	floatArg float = -10.2
print(floatArg + 1)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `-9.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeIntWithFlag(t *testing.T) {
	rsl := `
args:
	intArg int
print(intArg)
`
	setupAndRunCode(t, rsl, "--intArg", "-10")
	expected := `-10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeIntWithoutFlag(t *testing.T) {
	t.Skip("TODO: RAD-71") // todo RAD-71
	rsl := `
args:
	intArg int
print(intArg)
`
	setupAndRunCode(t, rsl, "-10")
	expected := `-10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeFloatWithFlag(t *testing.T) {
	rsl := `
args:
	floatArg float
print(floatArg)
`
	setupAndRunCode(t, rsl, "--floatArg", "-10.2")
	expected := `-10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanPassNegativeFloatWithoutFlag(t *testing.T) {
	t.Skip("TODO: RAD-71") // todo RAD-71
	rsl := `
args:
	floatArg float
print(floatArg)
`
	setupAndRunCode(t, rsl, "-10.2")
	expected := `-10.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanHaveSeveralMinuses(t *testing.T) {
	rsl := `
args:
	intArg int = --- 10
	floatArg float = -------10.2
print(intArg + 1)
print(floatArg + 1)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `-9
-9.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CanHaveIntAsDefaultForFloatArg(t *testing.T) {
	rsl := `
args:
	floatArg float = 2
print(floatArg + 1.2)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `3.2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestArgs_CannotHaveFloatAsDefaultForIntArg(t *testing.T) {
	rsl := `
args:
	intArg int = 1.2
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:16

  	intArg int = 1.2
                 ^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func TestArgs_FullHelp(t *testing.T) {
	rsl := `
args:
	name string # The name.
`
	setupAndRunCode(t, rsl, "--help", "--color=never")
	expected := `Usage:
  <name>

Script args:
      --name string   The name.

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_ShortHelp(t *testing.T) {
	rsl := `
args:
	name string # The name.
`
	setupAndRunCode(t, rsl, "-h", "--color=never")
	expected := `Usage:
  <name>

Script args:
      --name string   The name.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_ShortHelpNoArgs(t *testing.T) {
	rsl := `
print("hi")
`
	setupAndRunCode(t, rsl, "-h", "--color=never")
	expected := `Usage:
 
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_FullHelpNoArgs(t *testing.T) {
	rsl := `
print("hi")
`
	setupAndRunCode(t, rsl, "--help", "--color=never")
	expected := `Usage:
 

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_HelpWorksForAllTypes(t *testing.T) {
	rsl := `
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
	setupAndRunCode(t, rsl, "-h", "--color=never")
	expected := `Usage:
  [stringArg] [intArg] [floatArg] [boolArg] [stringArrayArg] [intArrayArg] [floatArrayArg] [boolArrayArg]

Script args:
      --stringArg string                (default alice)
      --intArg int                     An int. (default 1)
      --floatArg float                  (default 1.1)
      --boolArg                         (default true)
      --stringArrayArg string,string    (default ["bob", "charlie"])
      --intArrayArg int,int             (default [2, 3])
      --floatArrayArg float,float       (default [2.1, 3.1])
      --boolArrayArg bool,bool          (default [true, false])
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_UnsetBoolDefaultsToFalse(t *testing.T) {
	rsl := `
args:
	name string
	isTall bool
print(name, isTall)
`
	setupAndRunCode(t, rsl, "alice")
	expected := `alice false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_CanDefaultBoolToTrue(t *testing.T) {
	rsl := `
args:
	name string
	isTall bool = true
print(name, isTall)
`
	setupAndRunCode(t, rsl, "alice")
	expected := `alice true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
}

func TestArgs_MissingArgsPrintsUsageAndReturnsError(t *testing.T) {
	rsl := `
args:
	name string
	age int
`
	setupAndRunCode(t, rsl, "alice", "--color=never")
	expected := `Missing required arguments: [age]

Usage:
  <name> <age>

Script args:
      --name string   
      --age int       

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func TestArgs_TooManyArgsPrintsUsageAndReturnsError(t *testing.T) {
	rsl := `
args:
	name string
	age int
`
	setupAndRunCode(t, rsl, "alice", "2", "3", "--color=never")
	expected := `Too many positional arguments. Unused: [3]

Usage:
  <name> <age>

Script args:
      --name string   
      --age int       

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}

func TestArgs_InvalidFlagPrintsUsageAndReturnsError(t *testing.T) {
	rsl := `
args:
	name string
	age int
`
	setupAndRunCode(t, rsl, "alice", "2", "-s", "--color=never")
	expected := `unknown shorthand flag: 's' in -s

Usage:
  <name> <age>

Script args:
      --name string   
      --age int       

` + scriptGlobalFlagHelp
	assertError(t, 1, expected)
}
