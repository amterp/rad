package testing

import "testing"

func Test_Shell_ExportsNothingIfNoVars(t *testing.T) {
	script := `
2 + 3
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOutput(t, stdOutBuffer, "")
	assertOutput(t, stdErrBuffer, "")
	assertNoErrors(t)
}

func Test_Shell_ExportsStrings(t *testing.T) {
	script := `
a = "alice"
b = "bob"
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOutput(t, stdOutBuffer, "a=\"alice\"\nb=\"bob\"\n")
	assertOutput(t, stdErrBuffer, "")
	assertNoErrors(t)
}

func Test_Shell_PrintsGoToStderr(t *testing.T) {
	script := `
a = "alice"
print('hi')
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertOutput(t, stdErrBuffer, "hi\n")
	assertNoErrors(t)
}

func Test_Shell_ErrorExitFuncPrintsShellExitAndDoesNotExportVars(t *testing.T) {
	t.Skip("TOOD RAD-319: if we exit inside a function its params get exported, not good.")
	script := `
exit(2)
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "exit 2\n")
	assertError(t, 2, "")
}

func Test_Shell_NonErrorExitFuncStillExportsVars(t *testing.T) {
	t.Skip("TOOD RAD-319: if we exit inside a function its params get exported, not good.")
	script := `
a = "alice"
exit()
b = "bob"
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertNoErrors(t)
}

func Test_Shell_DoesNotExportLambdas(t *testing.T) {
	script := `
a = 1
b = fn() 2
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=1\n")
	assertNoErrors(t)
}

func Test_Shell_DoesNotExportNulls(t *testing.T) {
	script := `
a = 1
b = null
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=1\n")
	assertNoErrors(t)
}

func Test_Shell_DoesNotExportPath(t *testing.T) {
	script := `
a = 1
PATH = "noo"
`
	setupAndRunCode(t, script, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=1\n")
	assertNoErrors(t)
}
