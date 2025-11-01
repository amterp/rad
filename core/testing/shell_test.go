package testing

import (
	"testing"

	"github.com/amterp/rad/core"
)

func Test_ShellCmd_NoModifiers(t *testing.T) {
	script := `$"echo hello"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo hello",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_SingleAssignment_CodeOnly(t *testing.T) {
	script := `code = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_TwoAssignments_Positional(t *testing.T) {
	script := `code, out = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: true,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_ThreeAssignments_Positional(t *testing.T) {
	script := `code, out, err = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: true,
		CaptureStderr: true,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_NamedAssignment_OnlyStdout(t *testing.T) {
	script := `stdout = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: true,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_NamedAssignment_OnlyStderr(t *testing.T) {
	script := `stderr = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: false,
		CaptureStderr: true,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_NamedAssignment_CodeAndStderr(t *testing.T) {
	script := `code, stderr = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: false,
		CaptureStderr: true,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_NamedAssignment_StdoutAndStderr(t *testing.T) {
	script := `stdout, stderr = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: true,
		CaptureStderr: true,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_MixedNaming_FallsBackToPositional(t *testing.T) {
	script := `stderr, myvar = $"echo test"`
	setupAndRunCode(t, script, "--color=never")

	// Mixed naming falls back to positional: stderr gets code (0), myvar gets stdout
	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: true,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_QuietModifier(t *testing.T) {
	script := `quiet $"echo hi"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo hi",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_ConfirmModifier(t *testing.T) {
	script := `confirm $"ls /"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "ls /",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     true,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_QuietAndConfirm(t *testing.T) {
	script := `quiet confirm $"make all"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "make all",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     true,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_MultipleCommands_StateIsolation(t *testing.T) {
	script := `
$"echo first"
quiet $"echo second"
confirm $"echo third"
quiet confirm $"echo fourth"
`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t,
		core.ShellInvocation{
			Command:       "echo first",
			CaptureStdout: false,
			CaptureStderr: false,
			IsQuiet:       false,
			IsConfirm:     false,
		},
		core.ShellInvocation{
			Command:       "echo second",
			CaptureStdout: false,
			CaptureStderr: false,
			IsQuiet:       true,
			IsConfirm:     false,
		},
		core.ShellInvocation{
			Command:       "echo third",
			CaptureStdout: false,
			CaptureStderr: false,
			IsQuiet:       false,
			IsConfirm:     true,
		},
		core.ShellInvocation{
			Command:       "echo fourth",
			CaptureStdout: false,
			CaptureStderr: false,
			IsQuiet:       true,
			IsConfirm:     true,
		},
	)
	assertNoErrors(t)
}
