package testing

import (
	"bytes"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"os"
	"rad/core"
	"testing"
	"time"
)

var (
	// stateful, reset for each test
	stdInBuffer  = new(bytes.Buffer)
	stdOutBuffer = new(bytes.Buffer)
	stdErrBuffer = new(bytes.Buffer)
	errorOrExit  = ErrorOrExit{}
	// dont need reset
	testCmdInput = newTestCmdInput()
)

type ErrorOrExit struct {
	exitCode *int
	panicMsg *string
}

func newTestCmdInput() core.CmdInput {
	testExitFunc := func(code int) {
		errorOrExit.exitCode = &code
		panic(fmt.Sprintf("Exit code: %d", code))
	}
	return core.CmdInput{
		RIo: &core.RadIo{
			StdIn:  stdInBuffer,
			StdOut: stdOutBuffer,
			StdErr: stdErrBuffer,
		},
		RExit:  &testExitFunc,
		RClock: core.NewFixedClock(2019, 12, 13, 14, 15, 16, 123123123, time.UTC),
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
	rootCmd := setupCmd(t, args...)
	defer func() {
		if r := recover(); r != nil {
			msg := fmt.Sprintf("%v", r)
			errorOrExit.panicMsg = &msg
		}
	}()
	err := rootCmd.Execute()
	assert.NoError(t, err, "Command should execute without Cobra error")
}

func setupCmd(t *testing.T, args ...string) *cobra.Command {
	t.Helper()

	os.Args = append([]string{os.Args[0]}, args...)

	rootCmd := core.NewRootCmd(testCmdInput)
	core.InitCmd(rootCmd)

	rootCmd.SetOut(stdOutBuffer)
	rootCmd.SetErr(stdErrBuffer)
	return rootCmd
}

func resetTestState() {
	stdInBuffer.Reset()
	stdOutBuffer.Reset()
	stdErrBuffer.Reset()
	errorOrExit = ErrorOrExit{}
	core.ResetGlobals()
}

func assertOnlyOutput(t *testing.T, buffer *bytes.Buffer, expected string) {
	assertExpected(t, buffer, expected)
	assertAllElseEmpty(t)
}

func assertExpected(t *testing.T, buffer *bytes.Buffer, expected string) {
	t.Helper()
	actual := buffer.String()
	assert.Equal(t, expected, actual, "Output should match expected value")
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
	if code != nil {
		t.Errorf("Expected no exit code, got %d", *code)
	}
	msg := errorOrExit.panicMsg
	if msg != nil {
		t.Errorf("Expected no panic, got %s", *msg)
	}
}
