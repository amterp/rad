package testing

import (
	"bytes"
	"fmt"
	"os"
	"rad/core"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const globalFlagHelp = `Global flags:
  -h, --help                   Print usage string.
  -D, --DEBUG                  Enables debug output. Intended for RSL script developers.
      --RAD-DEBUG              Enables Rad debug output. Intended for Rad developers.
      --NO-COLOR               Disable colorized output.
  -Q, --QUIET                  Suppresses some output.
      --SHELL                  Outputs shell/bash exports of variables, so they can be eval'd
  -V, --VERSION                Print rad version information.
      --STDIN script-name      Enables reading RSL from stdin, and takes a string arg to be treated as the 'script name'.
      --MOCK-RESPONSE string   Add mock response for json requests (pattern:filePath)
`

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
	panicMsg *string
}

func newRunnerInputInput() core.RunnerInput {
	testExitFunc := func(code int) {
		errorOrExit.exitCode = &code
		panic(fmt.Sprintf("Exit code: %d", code))
	}
	sleepFunc := func(duration time.Duration) {
		millisSlept = append(millisSlept, duration.Milliseconds())
	}
	return core.RunnerInput{
		RIo: &core.RadIo{
			StdIn:  stdInBuffer,
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

	args := tp.args
	if tp.rsl != "" {
		stdInBuffer.WriteString(tp.rsl)
		args = append([]string{"--STDIN", "test"}, tp.args...)
	}
	runner := setupRunner(t, args...)
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v", r)
			errorOrExit.panicMsg = &msg
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
	if code == nil || *code == 0 {
		return
	}
	t.Errorf("Expected no exit code, got %d.\nStderr: %s", *code, stdErrBuffer.String())
	msg := errorOrExit.panicMsg
	t.Errorf("Expected no panic, got %s", *msg)
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
