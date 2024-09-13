package testing

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"rad/core"
	"testing"
)

var (
	stdInBuffer  = new(bytes.Buffer)
	stdOutBuffer = new(bytes.Buffer)
	stdErrBuffer = new(bytes.Buffer)
	testRadIo    = core.RadIo{
		StdIn:  stdInBuffer,
		StdOut: stdOutBuffer,
		StdErr: stdErrBuffer,
	}
)

func setupAndRun(t *testing.T, args ...string) {
	t.Helper()

	// reset for each test
	stdInBuffer.Reset()
	stdOutBuffer.Reset()
	stdErrBuffer.Reset()

	//originalArgs := os.Args
	os.Args = append([]string{os.Args[0]}, args...)
	//defer func() {
	//	os.Args = originalArgs
	//}()

	rootCmd := core.NewRootCmd(testRadIo)
	core.InitCmd(rootCmd)

	rootCmd.SetOut(stdOutBuffer)
	rootCmd.SetErr(stdErrBuffer)

	//rootCmd.SetArgs(args)

	err := rootCmd.Execute()
	assert.NoError(t, err, "Command should execute without error")
}

func assertOnly(t *testing.T, buffer *bytes.Buffer, expected string) {
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
