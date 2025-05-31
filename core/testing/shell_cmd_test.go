package testing

import "testing"

func Test_ShellCmd_CanEcho(t *testing.T) {
	script := `
$!"echo hi"
print('hi2')
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `hi
hi2
`
	expectedStderr := `⚡️ Running: echo hi
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}
