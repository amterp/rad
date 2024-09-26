package testing

import "testing"

const (
	setupArgRsl = `
args:
    foo "bar" x string`
)

func TestArgApiRename(t *testing.T) {
	rsl := setupArgRsl + `
print(foo)
`
	setupAndRunCode(t, rsl, "hey")
	expected := `hey
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestArgApiRenameUsageString(t *testing.T) {
	setupAndRunCode(t, setupArgRsl, "-h")
	expected := `Usage:
  test <bar> [flags]

Flags:
  -x, --bar string
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestPrintsUsageWithoutErrorIfNoArgsPassedOneRequiredOneOptionalArg(t *testing.T) {
	rsl := `
args:
	mandatory string
	optional int = 10
`
	setupAndRunCode(t, rsl)
	expected := `Usage:
  test <mandatory> [optional] [flags]

Flags:
      --mandatory string   
      --optional int        (default 10)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestInvokesIfNoArgsPassedButAllArgsAreOptional(t *testing.T) {
	rsl := `
args:
	optionalS string?
	optionalI int = 10
print('hi')
`
	setupAndRunCode(t, rsl)
	expected := `hi
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestErrorsIfSomeRequiredArgsMissing(t *testing.T) {
	rsl := `
args:
	mandatory1 string
	mandatory2 string
	optional int = 10
print('hi')
`
	setupAndRunCode(t, rsl, "one")
	expectedStdout := `Usage:
  test <mandatory1> <mandatory2> [optional] [flags]

Flags:
      --mandatory1 string   
      --mandatory2 string   
      --optional int         (default 10)
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertError(t, 1, "Missing required arguments: [mandatory2]\n")
	resetTestState()
}
