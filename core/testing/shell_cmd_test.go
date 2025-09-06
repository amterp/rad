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
	expectedStderr := `⚡️ echo hi
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CanAssign(t *testing.T) {
	script := `
code, out, err = $!"echo -n hi"
print('hi2')
print(code, out, err, sep="|")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `hi2
0|hi|
`
	expectedStderr := `⚡️ echo -n hi
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}
