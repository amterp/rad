package testing

import (
	"testing"
)

func Test_IndentNormalization_CommandDescription(t *testing.T) {
	script := `
command foo:
    ---
    Run the
    foo command.
    ---
    calls fn():
        print("hi")
`
	setupAndRunCode(t, script, "foo", "--help", "--color=never")
	expected := `Run the
foo command.

Usage:
  foo [OPTIONS]

Global options:
  -h, --help            Print usage string.
  -r, --repl            Start interactive REPL mode.
  -d, --debug           Enables debug output. Intended for Rad script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never] (default auto)
  -q, --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_IndentNormalization_ScriptDescription(t *testing.T) {
	script := `
---
Script description
with multiple lines.
---
print("hi")
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Script description
with multiple lines.

Usage:
  TestCase [OPTIONS]

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_IndentNormalization_ArgumentComment(t *testing.T) {
	script := `
args:
    name str  # The user's full name.
print(name)
`
	setupAndRunCode(t, script, "--help", "--color=never")
	expected := `Usage:
  TestCase <name> [OPTIONS]

Script args:
      --name str   The user's full name.

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_IndentNormalization_MixedIndentation(t *testing.T) {
	script := `
command test:
    ---
    Line 1
        Indented line
    Back to baseline
    ---
    calls fn():
        print("ok")
`
	setupAndRunCode(t, script, "test", "--help", "--color=never")
	expected := `Line 1
    Indented line
Back to baseline

Usage:
  test [OPTIONS]

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_IndentNormalization_WithTabs(t *testing.T) {
	script := "command foo:\n\t---\n\tTabbed\n\tdescription.\n\t---\n\tcalls fn():\n\t\tprint(\"hi\")\n"
	setupAndRunCode(t, script, "foo", "--help", "--color=never")
	expected := `Tabbed
description.

Usage:
  foo [OPTIONS]

` + scriptGlobalFlagHelp
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
