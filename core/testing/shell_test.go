package testing

import "testing"

func Test_Shell_ExportsNothingIfNoVars(t *testing.T) {
	rsl := `
2 + 3
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOutput(t, stdOutBuffer, "")
	assertOutput(t, stdErrBuffer, "")
	assertNoErrors(t)
}

func Test_Shell_ExportsStrings(t *testing.T) {
	rsl := `
a = "alice"
b = "bob"
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOutput(t, stdOutBuffer, "a=\"alice\"\nb=\"bob\"\n")
	assertOutput(t, stdErrBuffer, "")
	assertNoErrors(t)
}

func Test_Shell_PrintsGoToStderr(t *testing.T) {
	rsl := `
a = "alice"
print('hi')
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertOutput(t, stdErrBuffer, "hi\n")
	assertNoErrors(t)
}

func Test_Shell_ErrorExitFuncPrintsShellExitAndDoesNotExportVars(t *testing.T) {
	rsl := `
exit(2)
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "exit 2\n")
	assertError(t, 2, "")
}

func Test_Shell_NonErrorExitFuncStillExportsVars(t *testing.T) {
	rsl := `
a = "alice"
exit()
b = "bob"
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertNoErrors(t)
}

func Test_Shell_DoesNotExportLambdas(t *testing.T) {
	rsl := `
a = 1
b = fn() 2
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=1\n")
	assertNoErrors(t)
}

func Test_Shell_DoesNotExportNulls(t *testing.T) {
	rsl := `
a = 1
b = null
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=1\n")
	assertNoErrors(t)
}
