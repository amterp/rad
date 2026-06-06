package testing

import (
	"errors"
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

func Test_ShellCmd_Confirm_Decline(t *testing.T) {
	script := `confirm $"echo hi"`
	decline := func(title, prompt string) (bool, error) { return false, nil }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(decline))

	// Declining doesn't run the command, but still surfaces as a (catchable)
	// command failure with exit code 1.
	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20000", "Command exited with code 1")
}

func Test_ShellCmd_Confirm_DeclineCaught(t *testing.T) {
	script := `
confirm $"echo hi" catch:
    print("declined")
`
	decline := func(title, prompt string) (bool, error) { return false, nil }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(decline))

	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertOnlyOutput(t, stdOutBuffer, "declined\n")
	assertNoErrors(t)
}

func Test_ShellCmd_Confirm_Abort(t *testing.T) {
	script := `confirm $"echo hi"`
	abort := func(title, prompt string) (bool, error) { return false, errors.New("user aborted") }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(abort))

	// Aborting the prompt (Ctrl-C / Esc) is a clean, catchable user-input error,
	// NOT an internal-bug crash. This is the regression test for RAD20042.
	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20010", "Shell command aborted", "user aborted")
}

func Test_ShellCmd_Confirm_AbortCaught(t *testing.T) {
	script := `
confirm $"echo hi" catch:
    print("aborted")
`
	abort := func(title, prompt string) (bool, error) { return false, errors.New("user aborted") }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(abort))

	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertOnlyOutput(t, stdOutBuffer, "aborted\n")
	assertNoErrors(t)
}

func Test_ShellCmd_Confirm_DeclineKeepsCaptures(t *testing.T) {
	script := `
code, out = confirm $"echo hi" catch:
    print("code={code} out=[{out}]")
`
	decline := func(title, prompt string) (bool, error) { return false, nil }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(decline))

	// Declining must still leave capture targets defined (empty), just like a
	// command that ran and exited non-zero. Regression: `out` was previously
	// left undefined, blowing up the catch block with an undefined-variable error.
	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertOnlyOutput(t, stdOutBuffer, "code=1 out=[]\n")
	assertNoErrors(t)
}

func Test_ShellCmd_ConfirmShellFlag_Abort(t *testing.T) {
	script := `$"echo hi"`
	abort := func(title, prompt string) (bool, error) { return false, errors.New("user aborted") }
	setupAndRun(t, NewTestParams(script, "--confirm-shell", "--color=never").ConfirmResponder(abort))

	// The --confirm-shell flag confirms every command via the same path.
	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20010", "Shell command aborted")
}
