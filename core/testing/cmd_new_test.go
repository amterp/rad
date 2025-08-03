package testing

import (
	"testing"

	com "github.com/amterp/rad/core/common"
)

func Test_Cmd_New_FailsIfExists(t *testing.T) {
	setupAndRunArgs(t, "new", "cmd_new_test.go", "--color=never")
	expected := `Path 'cmd_new_test.go' already exists
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Cmd_New_SucceedsIfNotExists(t *testing.T) {
	defer com.DeleteFileIfExists("does_not_exist")
	setupAndRunArgs(t, "new", "does_not_exist", "--color=never")
	expectedStdout := `does_not_exist is ready to go.
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	expectedStderr := `⚡️ Running: touch does_not_exist
⚡️ Running: chmod +x does_not_exist
`
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_Cmd_New_CreatedFileIsRunnable(t *testing.T) {
	defer com.DeleteFileIfExists("does_not_exist")
	setupAndRunArgs(t, "new", "does_not_exist", "--color=never")
	stdOutBuffer.Reset()
	stdErrBuffer.Reset()
	assertNoErrors(t)

	setupAndRunArgs(t, "does_not_exist", "Alex", "--color=never")
	expected := `Hello, Alex!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
