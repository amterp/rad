package testing

import (
	"testing"
)

func Test_Cmds_CanDefineMultipleCommandsWithOwnArgs(t *testing.T) {
	script := `
command greet:
    name str
    age int
    calls fn():
        print("Hello {name}, age {age}")

command status:
    code int
    message str
    calls fn():
        print("Status {code}: {message}")
`

	setupAndRunCode(t, script, "greet", "Alice", "30", "--color=never")
	expected := `Hello Alice, age 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "status", "200", "OK", "--color=never")
	expected = `Status 200: OK
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "greet", "--name=Bob", "--age=25", "--color=never")
	expected = `Hello Bob, age 25
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "status", "--code=404", "--message=NotFound", "--color=never")
	expected = `Status 404: NotFound
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Cmds_CmdArgCanAppearOnMultipleCommands(t *testing.T) {
	script := `
args:
    verbose v bool

command greet:
    name str
    calls fn():
        print("greet: {name}, verbose={verbose}")

command farewell:
    name str
    calls fn():
        print("farewell: {name}, verbose={verbose}")
`

	setupAndRunCode(t, script, "greet", "Alice", "--verbose", "--color=never")
	expected := `greet: Alice, verbose=true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "farewell", "Bob", "--verbose", "--color=never")
	expected = `farewell: Bob, verbose=true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "greet", "Charlie", "--color=never")
	expected = `greet: Charlie, verbose=false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Cmds_CanHaveScriptArgsAndCommandArgs(t *testing.T) {
	script := `
args:
    verbose v bool
    count int = 1

command process:
    input str
    calls fn():
        for i in range(count):
            print("Processing {input} (verbose={verbose})")
`

	setupAndRunCode(t, script, "process", "data.txt", "--color=never")
	expected := `Processing data.txt (verbose=false)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "process", "data.txt", "--verbose", "--count=3", "--color=never")
	expected = `Processing data.txt (verbose=true)
Processing data.txt (verbose=true)
Processing data.txt (verbose=true)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Cmds_Help(t *testing.T) {
	script := `
args:
    verbose v bool
    count int = 1  # A count.

command process:
    ---
    Process something.
    Second line!
    ---
    input str  # An input.
    calls fn():
        for i in range(count):
            print("Processing {input} (verbose={verbose})")
`

	setupAndRunCode(t, script, "-h", "--color=never")
	// TODO this should just say 'command', not 'subcommand'
	expected := `Usage:
  TestCase [subcommand] [OPTIONS]

Commands:
  process   Process something.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "--help", "--color=never")
	expected = `Usage:
  TestCase [subcommand] [OPTIONS]

Commands:
  process   Process something.

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "process", "-h", "--color=never")
	expected = `Process something.
Second line!

Usage:
  process <input> [count] [OPTIONS]

Command args:
      --input str   An input.
  -v, --verbose
      --count int   A count. (default 1)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	setupAndRunCode(t, script, "process", "--help", "--color=never")
	expected = `Process something.
Second line!

Usage:
  process <input> [count] [OPTIONS]

Command args:
      --input str   An input.
  -v, --verbose
      --count int   A count. (default 1)

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Cmds_ErrorsIfCommandNotSpecified(t *testing.T) {
	script := `
command greet:
    calls fn():
        print("Hello!")

command farewell:
    calls fn():
        print("Goodbye!")
`

	setupAndRunCode(t, script, "--color=never")
	expected := `Must specify a command

Usage:
  TestCase [subcommand] [OPTIONS]

Commands:
  farewell
  greet

` + scriptGlobalFlagHelp
	assertOutput(t, stdOutBuffer, "")
	assertError(t, 1, expected)
}

func Test_NoScript_ErrorsOnUnknownArguments(t *testing.T) {
	setupAndRunArgs(t, "--unknown-flag", "--color=never")
	expected := `Unknown arguments: [--unknown-flag]


` + radHelp
	assertOutput(t, stdOutBuffer, "")
	assertError(t, 1, expected)
}
