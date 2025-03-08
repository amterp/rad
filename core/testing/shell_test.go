package testing

import "testing"

func TestShell_ExportsNothingIfNoVars(t *testing.T) {
	rsl := `
2 + 3
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
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
	setupAndRunCode(t, rsl, "--color=never", "--shell")
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
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertOutput(t, stdErrBuffer, "hi\n")
	assertNoErrors(t)
	resetTestState()
}

func TestShell_ErrorExitFuncPrintsShellExitAndDoesNotExportVars(t *testing.T) {
	rsl := `
exit(2)
`
	setupAndRunCode(t, rsl, "--color=never", "--shell")
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
	setupAndRunCode(t, rsl, "--color=never", "--shell")
	assertOnlyOutput(t, stdOutBuffer, "a=\"alice\"\n")
	assertNoErrors(t)
	resetTestState()
}
