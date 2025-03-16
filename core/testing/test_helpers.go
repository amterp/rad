package testing

import (
	"bytes"
	"fmt"
	"os"
	"rad/core"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const scriptGlobalFlagHelp = `Global flags:
  -h, --help            Print usage string.
  -d, --debug           Enables debug output. Intended for RSL script developers.
      --color mode      Control output colorization. Valid values: [auto, always, never]. (default auto)
  -q, --quiet           Suppresses some output.
      --confirm-shell   Confirm all shell commands before running them.
      --src             Instead of running the target script, just print it out.
`

const allGlobalFlagHelp = `Global flags:
  -h, --help                   Print usage string.
  -d, --debug                  Enables debug output. Intended for RSL script developers.
      --rad-debug              Enables Rad debug output. Intended for Rad developers.
      --color mode             Control output colorization. Valid values: [auto, always, never]. (default auto)
  -q, --quiet                  Suppresses some output.
      --shell                  Outputs shell/bash exports of variables, so they can be eval'd
  -v, --version                Print rad version information.
      --confirm-shell          Confirm all shell commands before running them.
      --src                    Instead of running the target script, just print it out.
      --rsl-tree               Instead of running the target script, print out its syntax tree.
      --mock-response string   Add mock response for json requests (pattern:filePath)
`

const radHelp = `rad: A tool for writing user-friendly command line scripts.
GitHub: https://github.com/amterp/rad
Documentation: https://amterp.github.io/rad/

Usage:
  rad [script path | command] [flags]

Commands:
  new           Sets up a new RSL script.

To see help for a specific command, run ` + "`rad <command> -h`.\n\n" + allGlobalFlagHelp + `
To execute an RSL script:
  rad path/to/script.rsl [args]

To execute a command:
  rad <command> [args]

If you're new, check out the Getting Started guide: https://amterp.github.io/rad/guide/getting-started/
`

const ignorePanicMsg = "TESTING - IGNORE ME"

var (
	// stateful, reset for each test
	stdInBuffer  = new(bytes.Buffer)
	stdOutBuffer = new(bytes.Buffer)
	stdErrBuffer = new(bytes.Buffer)
	errorOrExit  = ErrorOrExit{}
	millisSlept  = make([]int64, 0)
	// dont need reset
	runnerInputInput = newRunnerInputInput()
)

type ErrorOrExit struct {
	exitCode *int
	//stderrSnapshot string
	panicMsg *string
}

func newRunnerInputInput() core.RunnerInput {
	testExitFunc := func(code int) {
		errorOrExit.exitCode = &code
		//errorOrExit.stderrSnapshot = stdErrBuffer.String()
		panic(ignorePanicMsg)
	}
	sleepFunc := func(duration time.Duration) {
		millisSlept = append(millisSlept, duration.Milliseconds())
	}
	return core.RunnerInput{
		RIo: &core.RadIo{
			StdIn:  core.NewBufferReader(stdInBuffer),
			StdOut: stdOutBuffer,
			StdErr: stdErrBuffer,
		},
		RExit:  &testExitFunc,
		RClock: core.NewFixedClock(2019, 12, 13, 14, 15, 16, 123123123, time.UTC),
		RSleep: &sleepFunc,
	}
}

type TestParams struct {
	rsl        string
	stdinInput string // todo not implemented
	args       []string
}

func NewTestParams(rsl string, args ...string) *TestParams {
	return &TestParams{
		rsl:  rsl,
		args: args,
	}
}

func (tp *TestParams) StdinInput(stdinInput string) *TestParams {
	tp.stdinInput = stdinInput
	return tp
}

func setupAndRunCode(t *testing.T, rsl string, args ...string) {
	setupAndRun(t, NewTestParams(rsl, args...))
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
	if tp.rsl != "" {
		stdInBuffer.WriteString(tp.rsl)
		args = append([]string{"-"}, tp.args...)
	}
	runner := setupRunner(t, args...)
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v", r)
			if !strings.Contains(msg, ignorePanicMsg) {
				//errorOrExit.stderrSnapshot += msg
				errorOrExit.panicMsg = &msg
			}
		}
	}()
	err := runner.Run()
	assert.NoError(t, err, "Command should execute without error")
}

func setupRunner(t *testing.T, args ...string) *core.RadRunner {
	t.Helper()

	os.Args = append([]string{os.Args[0]}, args...)

	return core.NewRadRunner(runnerInputInput)
}

func resetTestState() {
	stdInBuffer.Reset()
	stdOutBuffer.Reset()
	stdErrBuffer.Reset()
	errorOrExit = ErrorOrExit{}
	millisSlept = make([]int64, 0)
	core.ResetGlobals()
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
}

func assertError(t *testing.T, expectedCode int, expectedMsg string) {
	t.Helper()
	//stdErrBuffer.Reset()
	//errBuffer := bytes.NewBufferString(errorOrExit.stderrSnapshot)
	//errorOrExit.stderrSnapshot = ""
	assertOnlyOutput(t, stdErrBuffer, expectedMsg)
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
