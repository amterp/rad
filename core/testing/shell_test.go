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

func Test_ShellCmd_QuietModifier_SuppressesOutput(t *testing.T) {
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

func Test_ShellCmd_ConfirmModifier_IsDetected(t *testing.T) {
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

func Test_ShellCmd_QuietAndConfirm_Together(t *testing.T) {
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

func Test_ShellCmd_ConfirmAndQuiet_OrderReversed(t *testing.T) {
	script := `confirm quiet $"make clean"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "make clean",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     true,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_Quiet_WithPositionalAssignment(t *testing.T) {
	script := `code, stdout = quiet $"echo output"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo output",
		CaptureStdout: true,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_Quiet_WithNamedAssignment(t *testing.T) {
	script := `stdout, code = quiet $"echo output"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo output",
		CaptureStdout: true,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_Quiet_WithThreeAssignments(t *testing.T) {
	script := `code, stdout, stderr = quiet $"echo error >&2"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo error >&2",
		CaptureStdout: true,
		CaptureStderr: true,
		IsQuiet:       true,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_Quiet_WithCatchBlock(t *testing.T) {
	script := `
quiet $"exit 1" catch:
    print("caught")
`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "exit 1",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     false,
	})

	// The mock returns exit code 0, so catch block won't execute in this test
	// But the modifier parsing is what we're testing
	assertNoErrors(t)
}

func Test_ShellCmd_Confirm_WithCatchBlock(t *testing.T) {
	script := `
confirm $"exit 1" catch:
    print("error handled")
`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "exit 1",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     true,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_MultipleCommands_DifferentModifiers(t *testing.T) {
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

func Test_ShellCmd_Quiet_WithStringInterpolation(t *testing.T) {
	script := `
name = "test"
quiet $"echo {name}"
`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "echo test",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     false,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_Confirm_WithComplexCommand(t *testing.T) {
	script := `confirm $"ls -la | grep test | wc -l"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "ls -la | grep test | wc -l",
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       false,
		IsConfirm:     true,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_QuietConfirm_WithComplexCommand(t *testing.T) {
	script := `quiet confirm $"find . -name '*.go' | xargs rm"`
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       `find . -name '*.go' | xargs rm`,
		CaptureStdout: false,
		CaptureStderr: false,
		IsQuiet:       true,
		IsConfirm:     true,
	})
	assertNoErrors(t)
}
