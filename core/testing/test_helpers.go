package testing

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/amterp/color"
	"github.com/amterp/rad/core"

	"github.com/stretchr/testify/assert"
)

const scriptGlobalFlagHelp = `Global options:
  -h, --help            Print usage string.
  -r, --repl            Start interactive REPL mode.
  -d, --debug           Enables debug output. Intended for Rad script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never] (default auto)
  -q, --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`

const allGlobalFlagHelp = `Global options:
  -h, --help                Print usage string.
  -r, --repl                Start interactive REPL mode.
  -d, --debug               Enables debug output. Intended for Rad script developers.
      --rad-debug           Enables Rad debug output. Intended for Rad developers.
      --color mode          Control output colorization. Valid values: [auto, always, never] (default auto)
  -q, --quiet               Suppresses some output.
      --shell               Outputs shell/bash exports of variables, so they can be eval'd
  -v, --version             Print rad version information.
      --confirm-shell       Confirm all shell commands before running them.
      --src                 Instead of running the target script, just print it out.
      --cst-tree            Instead of running the target script, print out its CST (concrete syntax tree).
      --ast-tree            Instead of running the target script, print out its AST (abstract syntax tree).
      --rad-args-dump       Instead of running the target script, print out an args dump for debugging argument parsing.
      --mock-response str   (optional) Add mock response for json requests (pattern:filePath)
`

const radHelp = `rad: A tool for writing user-friendly command line scripts.
GitHub: https://github.com/amterp/rad
Documentation: https://amterp.github.io/rad/

Usage:
  rad [script path | command] [flags]

Commands:
  new           Sets up a new Rad script.
  docs          Opens rad's documentation website.
  check         Validates & lints Rad scripts.
  home          Prints out rad's home directory.
  gen-id        Generates a unique string ID. Useful for e.g. rad stash IDs.
  stash         Interacts with script stashes.
  explain       Explains Rad error codes with detailed documentation.

To see help for a specific command, run ` + "`rad <command> -h`.\n\n" + allGlobalFlagHelp + `
To execute a Rad script:
  rad path/to/script.rad [args]

To execute a command:
  rad <command> [args]

If you're new, check out the Getting Started guide: https://amterp.github.io/rad/guide/getting-started/
`

const ignorePanicMsg = "TESTING - IGNORE ME"

var (
	// stateful, reset for each test
	stdInBuffer      = new(bytes.Buffer)
	stdOutBuffer     = new(bytes.Buffer)
	stdErrBuffer     = new(bytes.Buffer)
	errorOrExit      = ErrorOrExit{}
	millisSlept      = make([]int64, 0)
	shellInvocations = make([]core.ShellInvocation, 0)
	httpInvocations  = make([]core.HttpRequest, 0)
	runnerInput      = newRunnerInput()
)

type ErrorOrExit struct {
	exitCode *int
	// stderrSnapshot string
	panicMsg *string
}

func newRunnerInput() core.RunnerInput {
	testExitFunc := func(code int) {
		errorOrExit.exitCode = &code
		// errorOrExit.stderrSnapshot = stdErrBuffer.String()
		panic(ignorePanicMsg)
	}
	sleepFunc := func(duration time.Duration) {
		millisSlept = append(millisSlept, duration.Milliseconds())
	}
	shellExec := func(invocation core.ShellInvocation) (string, string, int) {
		shellInvocations = append(shellInvocations, invocation)
		// Return empty strings and exit code 0 for test mock
		return "", "", 0
	}
	requester := core.NewRequester()
	requester.SetCaptureCallback(func(inv core.HttpRequest) {
		httpInvocations = append(httpInvocations, inv)
	})
	radTestHome := filepath.Join("./rad_test_home")
	return core.RunnerInput{
		RIo: &core.RadIo{
			StdIn:  core.NewBufferReader(stdInBuffer),
			StdOut: stdOutBuffer,
			StdErr: stdErrBuffer,
		},
		RExit:   &testExitFunc,
		RClock:  core.NewFixedClock(2019, 12, 13, 14, 15, 16, 123123123, time.UTC),
		RSleep:  &sleepFunc,
		RShell:  &shellExec,
		RReq:    requester,
		RadHome: &radTestHome,
	}
}

type TestParams struct {
	script        string
	stdinInput    string
	stdinInputSet bool
	args          []string
}

func NewTestParams(script string, args ...string) *TestParams {
	return &TestParams{
		script: script,
		args:   args,
	}
}

func (tp *TestParams) StdinInput(stdinInput string) *TestParams {
	tp.stdinInput = stdinInput
	tp.stdinInputSet = true
	return tp
}

func setupAndRunCode(t *testing.T, script string, args ...string) {
	setupAndRun(t, NewTestParams(script, args...))
}

func setupAndRunArgs(t *testing.T, args ...string) {
	t.Helper()
	setupAndRun(t, NewTestParams("", args...))
}

func setupAndRun(t *testing.T, tp *TestParams) {
	t.Helper()
	resetTestState()
	core.IsTest = true

	args := tp.args

	// Handle different combinations of script and stdinInput
	if tp.script != "" && tp.stdinInputSet {
		// Both script and stdin data provided - write script to temp file
		tmpFile, err := os.CreateTemp("", "rad_test_*.rad")
		if err != nil {
			t.Fatalf("Failed to create temp file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(tp.script); err != nil {
			t.Fatalf("Failed to write script to temp file: %v", err)
		}
		tmpFile.Close()

		// Write stdin data to buffer and mark as piped
		stdInBuffer.WriteString(tp.stdinInput)
		if br, ok := runnerInput.RIo.StdIn.(*core.BufferReader); ok {
			br.SetPiped(true)
		}

		// Run the temp file
		args = append([]string{tmpFile.Name()}, tp.args...)
	} else if tp.script != "" {
		// Only script provided - use stdin for script (rad -)
		// Don't set isPiped here - stdin is consumed by reading the script,
		// not available for the script to read
		stdInBuffer.WriteString(tp.script)
		args = append([]string{"-"}, tp.args...)
	} else if tp.stdinInputSet {
		// Only stdin data provided - write to buffer and mark as piped
		stdInBuffer.WriteString(tp.stdinInput)
		if br, ok := runnerInput.RIo.StdIn.(*core.BufferReader); ok {
			br.SetPiped(true)
		}
	}

	// Set NO_COLOR for tests that pass --color=never
	// Early validation happens before --color flag is parsed, so we need env var
	hasColorNever := false
	for _, arg := range args {
		if arg == "--color=never" || (strings.HasPrefix(arg, "--color") && strings.Contains(arg, "never")) {
			hasColorNever = true
			break
		}
	}
	if hasColorNever {
		os.Setenv("NO_COLOR", "1")
	} else {
		os.Unsetenv("NO_COLOR")
	}

	runner := setupRunner(t, args...)
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v", r)
			if !strings.Contains(msg, ignorePanicMsg) {
				// errorOrExit.stderrSnapshot += msg
				errorOrExit.panicMsg = &msg
			}
		}
	}()
	err := runner.Run()
	assert.NoError(t, err, "Command should execute without error")
}

func setupRunner(t *testing.T, args ...string) *core.RadRunner {
	t.Helper()

	os.Args = append([]string{"rad"}, args...)

	return core.NewRadRunner(runnerInput)
}

func resetTestState() {
	stdInBuffer.Reset()
	stdOutBuffer.Reset()
	stdErrBuffer.Reset()
	errorOrExit = ErrorOrExit{}
	millisSlept = make([]int64, 0)
	shellInvocations = make([]core.ShellInvocation, 0)
	httpInvocations = make([]core.HttpRequest, 0)
	core.ResetGlobals()
	// ResetGlobals sets color.NoColor = false (production default).
	// Tests should default to no color; individual tests opt in via --color=always.
	color.NoColor = true
	// Clear mock patterns to prevent test interference
	runnerInput.RReq.ClearMockedResponses()
	// Reset the isPiped flag for stdin
	if br, ok := runnerInput.RIo.StdIn.(*core.BufferReader); ok {
		br.SetPiped(false)
	}
	// Reset NO_COLOR environment variable
	os.Unsetenv("NO_COLOR")
}

func assertOnlyOutput(t *testing.T, buffer *bytes.Buffer, expected string) {
	assertOutput(t, buffer, expected)
	assertAllElseEmpty(t)
}

func assertOutput(t *testing.T, buffer *bytes.Buffer, expected string) {
	t.Helper()
	actual := buffer.String()
	ok := assert.Equal(t, expected, actual, "Output should match expected value")
	if !ok {
		stderr := stdErrBuffer.String()
		if stderr != "" {
			t.Errorf("Stderr: %s", stderr)
		} else {
			t.Errorf("Stderr was empty")
		}
	}
	buffer.Reset()
}

func assertAllElseEmpty(t *testing.T) {
	t.Helper()
	assert.Empty(t, stdOutBuffer.String(), "Expected no output on stdout")
	assert.Empty(t, stdErrBuffer.String(), "Expected no output on stderr")
	assertNoHttpInvocations(t)
}

func assertError(t *testing.T, expectedCode int, expectedMsg string) {
	t.Helper()
	assertOnlyOutput(t, stdErrBuffer, expectedMsg)
	assertExitCode(t, expectedCode)
}

// assertErrorContains verifies that an error occurred with the expected code
// and that the error message contains all the given substrings. This is useful
// for platform-independent error testing where OS-specific messages may differ.
func assertErrorContains(t *testing.T, expectedCode int, substrings ...string) {
	t.Helper()
	actual := stdErrBuffer.String()
	for _, substr := range substrings {
		if !strings.Contains(actual, substr) {
			t.Errorf("Expected stderr to contain %q, but got:\n%s", substr, actual)
		}
	}
	stdErrBuffer.Reset()
	assertExitCode(t, expectedCode)
}

func assertExitCode(t *testing.T, code int) {
	t.Helper()
	assert.Equal(t, code, *errorOrExit.exitCode)
}

func assertNoErrors(t *testing.T) {
	t.Helper()
	code := errorOrExit.exitCode

	if code != nil && *code != 0 {
		t.Errorf("Expected no exit code, got %d.\nStderr: %s", *code, stdErrBuffer.String())
	}

	if errorOrExit.panicMsg != nil {
		t.Errorf("Expected no panic, got %s", *errorOrExit.panicMsg)
	}
}

func assertDidNotSleep(t *testing.T) {
	assertSleptMillis(t) // providing no millis
}

func assertSleptMillis(t *testing.T, millis ...int64) {
	if len(millisSlept) != len(millis) {
		t.Errorf("Expected to sleep %d times, but slept only %d times: %v", len(millis), len(millisSlept), millisSlept)
	}

	for i, expected := range millis {
		actual := millisSlept[i]
		if actual != expected {
			t.Errorf("Expected to sleep idx %d to be %d millis, but slept %d millis", i, expected, actual)
		}
	}
}

// Shell command assertion helpers

func assertShellNotInvoked(t *testing.T) {
	t.Helper()
	if len(shellInvocations) != 0 {
		t.Errorf("Expected no shell commands, but got %d invocations: %v", len(shellInvocations), shellInvocations)
	}
}

func assertShellCount(t *testing.T, count int) {
	t.Helper()
	if len(shellInvocations) != count {
		t.Errorf("Expected %d shell invocations, but got %d: %v", count, len(shellInvocations), shellInvocations)
	}
}

func assertShellInvoked(t *testing.T, expected ...core.ShellInvocation) {
	t.Helper()
	if len(shellInvocations) != len(expected) {
		t.Errorf("Expected %d shell invocations, but got %d.\nExpected: %+v\nActual: %+v",
			len(expected), len(shellInvocations), expected, shellInvocations)
		return
	}

	for i, exp := range expected {
		actual := shellInvocations[i]
		if actual.Command != exp.Command {
			t.Errorf("Invocation %d: Expected command %q, but got %q", i, exp.Command, actual.Command)
		}
		if actual.CaptureStdout != exp.CaptureStdout {
			t.Errorf("Invocation %d: Expected CaptureStdout=%v, but got %v", i, exp.CaptureStdout, actual.CaptureStdout)
		}
		if actual.CaptureStderr != exp.CaptureStderr {
			t.Errorf("Invocation %d: Expected CaptureStderr=%v, but got %v", i, exp.CaptureStderr, actual.CaptureStderr)
		}
		if actual.IsQuiet != exp.IsQuiet {
			t.Errorf("Invocation %d: Expected IsQuiet=%v, but got %v", i, exp.IsQuiet, actual.IsQuiet)
		}
		if actual.IsConfirm != exp.IsConfirm {
			t.Errorf("Invocation %d: Expected IsConfirm=%v, but got %v", i, exp.IsConfirm, actual.IsConfirm)
		}
	}
}

func assertNoHttpInvocations(t *testing.T) {
	t.Helper()
	if len(httpInvocations) != 0 {
		t.Errorf("Expected no HTTP invocations, but got %d", len(httpInvocations))
	}
}

func assertHttpInvocationUrls(t *testing.T, expectedUrls ...string) {
	t.Helper()
	if len(httpInvocations) != len(expectedUrls) {
		t.Errorf("Expected %d HTTP invocations, got %d", len(expectedUrls), len(httpInvocations))
		return
	}
	for i, expected := range expectedUrls {
		if httpInvocations[i].RequestDef.Url != expected {
			t.Errorf("HTTP invocation %d: expected URL %s, got %s",
				i, expected, httpInvocations[i].RequestDef.Url)
		}
	}
}
