package testing

import (
	com "rad/core/common"
	"testing"
)

func Test_Cmd_New_FailsIfExists(t *testing.T) {
	setupAndRunArgs(t, "new", "cmd_new_test.go", "--color=never")
	expected := `Path 'cmd_new_test.go' already exists
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Cmd_New_SucceedsIfNotExists(t *testing.T) {
	setupAndRunArgs(t, "new", "does_not_exist", "--color=never")
	expectedStdout := `does_not_exist is ready to go.
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	expectedStderr := `⚡️ Running: touch does_not_exist
⚡️ Running: chmod +x does_not_exist
`
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
	com.DeleteFileIfExists("does_not_exist")
}
