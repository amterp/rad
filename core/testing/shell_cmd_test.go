package testing

import "testing"

func Test_ShellCmd_CanEcho(t *testing.T) {
	script := `
$"echo hi"
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
code, out, err = $"echo -n hi"
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

func Test_ShellCmd_CriticalByDefault_FailsScript(t *testing.T) {
	script := `
$"exit 1"
print("Should not reach here")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStderr := `⚡️ exit 1
Error at L2:1

  $"exit 1"
  ^^^^^^^^^ Command exited with code 1
`
	assertOutput(t, stdOutBuffer, "")
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertExitCode(t, 1)
}

func Test_ShellCmd_SuccessfulCommand_Continues(t *testing.T) {
	script := `
$"exit 0"
print("Continued execution")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Continued execution
`
	expectedStderr := `⚡️ exit 0
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CatchBlock_NoAssignment(t *testing.T) {
	script := `
$"exit 1" catch:
	print("Caught error")
print("Continued execution")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Caught error
Continued execution
`
	expectedStderr := `⚡️ exit 1
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CatchBlock_WithAssignment(t *testing.T) {
	script := `
code, out = $"exit 5" catch:
	print("Caught error with code: {code}")
print("Final code: {code}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Caught error with code: 5
Final code: 5
`
	expectedStderr := `⚡️ exit 5
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CatchBlock_CanReassign(t *testing.T) {
	script := `
code = $"exit 1" catch:
	code = 0
print("Code: {code}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Code: 0
`
	expectedStderr := `⚡️ exit 1
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CatchBlock_ReturnInCatch(t *testing.T) {
	script := `
fn test():
	for i in range(10):
		code = $"exit 1" catch:
			return "early return"
	return "completed"

result = test()
print("Result: {result}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Result: early return
`
	expectedStderr := `⚡️ exit 1
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CatchBlock_BreakInCatch(t *testing.T) {
	script := `
for i in range(10):
	$"exit 1" catch:
		print("Breaking at {i}")
		break
	print("After shell command")

print("After loop")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Breaking at 0
After loop
`
	expectedStderr := `⚡️ exit 1
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CatchBlock_ContinueInCatch(t *testing.T) {
	script := `
for i in range(3):
	print("Iteration {i}")
	$"exit 1" catch:
		print("Continuing...")
		continue
	print("After shell (not printed)")

print("Done")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Iteration 0
Continuing...
Iteration 1
Continuing...
Iteration 2
Continuing...
Done
`
	expectedStderr := `⚡️ exit 1
⚡️ exit 1
⚡️ exit 1
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_PositionalAssignment(t *testing.T) {
	script := `
c, out = $"printf hello"
print("c={c}, out={out}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `c=0, out=hello
`
	expectedStderr := `⚡️ printf hello
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_NamedAssignment_StdoutCode(t *testing.T) {
	script := `
stdout, code = $"printf hello"
print("code={code}, stdout={stdout}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `code=0, stdout=hello
`
	expectedStderr := `⚡️ printf hello
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_NamedAssignment_CodeStderr(t *testing.T) {
	script := `
code, stderr = $"sh -c '>&2 printf error; exit 0'"
print("code={code}, stderr={stderr}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `code=0, stderr=error
`
	expectedStderr := `⚡️ sh -c '>&2 printf error; exit 0'
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_MixedNaming_UsesPositional(t *testing.T) {
	script := `
stdout, myvar = $"printf hello"
print("stdout={stdout}, myvar={myvar}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `stdout=0, myvar=hello
`
	expectedStderr := `⚡️ printf hello
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CaptureCode_Only(t *testing.T) {
	script := `
c = $"echo visible"
print("Code: {c}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `visible
Code: 0
`
	expectedStderr := `⚡️ echo visible
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CaptureTwoVars_StderrToTerminal(t *testing.T) {
	script := `
code, out = $"sh -c '>&2 echo to-stderr; printf to-stdout'"
print("Out: {out}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Out: to-stdout
`
	expectedStderr := `⚡️ sh -c '>&2 echo to-stderr; printf to-stdout'
to-stderr
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}

func Test_ShellCmd_CaptureThreeVars_NothingToTerminal(t *testing.T) {
	script := `
code, out, err = $"sh -c '>&2 printf to-stderr; printf to-stdout'"
print("Out: {out}, Err: {err}")
`
	setupAndRunCode(t, script, "--color=never")
	expectedStdout := `Out: to-stdout, Err: to-stderr
`
	expectedStderr := `⚡️ sh -c '>&2 printf to-stderr; printf to-stdout'
`
	assertOutput(t, stdOutBuffer, expectedStdout)
	assertOutput(t, stdErrBuffer, expectedStderr)
	assertNoErrors(t)
}
