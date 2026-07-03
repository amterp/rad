package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Under --shell, stdout is reserved for eval-able output. Help must land on
// stderr, with an `exit 0` on stdout so an eval'ing wrapper stops after
// showing help. See https://github.com/amterp/rad/issues/147.

const shellModeScript = `args:
    env str
`

func Test_ShellMode_Help_UsageOnStderr_ExitZeroOnStdout(t *testing.T) {
	setupAndRunCode(t, shellModeScript, "--shell", "--help", "--color=never")
	assertOutput(t, stdOutBuffer, "exit 0\n")
	assertErrorContains(t, 0, "Usage:", "--env")
}

func Test_ShellMode_Help_FlagOrderIrrelevant(t *testing.T) {
	setupAndRunCode(t, shellModeScript, "--help", "--shell", "--color=never")
	assertOutput(t, stdOutBuffer, "exit 0\n")
	assertErrorContains(t, 0, "Usage:", "--env")
}

func Test_ShellMode_HelpWithoutShell_UsageStaysOnStdout(t *testing.T) {
	setupAndRunCode(t, shellModeScript, "--help", "--color=never")
	assert.Contains(t, stdOutBuffer.String(), "Usage:")
	assert.Contains(t, stdOutBuffer.String(), "--env")
	assert.Empty(t, stdErrBuffer.String())
	assertExitCode(t, 0)
	stdOutBuffer.Reset()
}

func Test_ShellMode_MissingRequiredArg_EmitsShellExitOnStdout(t *testing.T) {
	setupAndRunCode(t, shellModeScript, "--shell", "--color=never")
	assertOutput(t, stdOutBuffer, "exit 1\n")
	assertErrorContains(t, 1, "Missing required arguments: [env]")
}

func Test_ShellMode_InvalidArg_EmitsShellExitOnStdout(t *testing.T) {
	setupAndRunCode(t, shellModeScript, "--shell", "dev", "extra", "--color=never")
	assertOutput(t, stdOutBuffer, "exit 1\n")
	assertErrorContains(t, 1, "Too many positional arguments")
}

func Test_ShellMode_RuntimeError_EmitsShellExitOnStdout(t *testing.T) {
	script := `a = [1]
print(a[5])
`
	setupAndRunCode(t, script, "--shell", "--color=never")
	assertOutput(t, stdOutBuffer, "exit 1\n")
	assertErrorContains(t, 1, "out of bounds")
}
