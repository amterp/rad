package testing

import "testing"

func TestShell_ExportsNothingIfNoVars(t *testing.T) {
	rsl := `
`
	setupAndRunCode(t, rsl, "--COLOR=never", "--SHELL")
	assertOutput(t, stdOutBuffer, "")
	assertOutput(t, stdErrBuffer, "")
	assertNoErrors(t)
	resetTestState()
}

func TestShell_ExportsStrings(t *testing.T) {
	rsl := `
a = "alice"
b = "bob"
`
	setupAndRunCode(t, rsl, "--COLOR=never", "--SHELL")
	assertOutput(t, stdOutBuffer, "a=\"alice\"\nb=\"bob\"\n")
	assertOutput(t, stdErrBuffer, "")
	assertNoErrors(t)
	resetTestState()
}

func TestShell_PrintsGoToStderr(t *testing.T) {
	rsl := `
a = "alice"
print('hi')
`
	setupAndRunCode(t, rsl, "--COLOR=never", "--SHELL")
	assertOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertOutput(t, stdErrBuffer, "hi\n")
	assertNoErrors(t)
	resetTestState()
}

func TestShell_ErrorExitFuncPrintsShellExitAndDoesNotExportVars(t *testing.T) {
	rsl := `
exit(2)
`
	setupAndRunCode(t, rsl, "--COLOR=never", "--SHELL")
	assertOnlyOutput(t, stdOutBuffer, "exit 2\n")
	assertError(t, 2, "")
	resetTestState()
}

func TestShell_NonErrorExitFuncStillExportsVars(t *testing.T) {
	rsl := `
a = "alice"
exit()
b = "bob"
`
	setupAndRunCode(t, rsl, "--COLOR=never", "--SHELL")
	assertOnlyOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertNoErrors(t)
	resetTestState()
}
