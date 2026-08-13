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

func Test_ShellCmd_SingleAssignment_NamedCodeCapturesNothing(t *testing.T) {
	// `code` is a reserved name, so this is named assignment: the exit code
	// needs no capture, and both streams still reach the terminal.
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

func Test_ShellCmd_OneAssignment_Positional(t *testing.T) {
	script := `out = $"echo test"`
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

func Test_ShellCmd_TwoAssignments_Positional(t *testing.T) {
	script := `out, err = $"echo test"`
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

func Test_ShellCmd_ThreeAssignments_Positional(t *testing.T) {
	script := `out, err, c = $"echo test"`
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

// The next three tests pin the positional *order* - (stdout, stderr, code) -
// rather than just which streams got captured. They need distinguishable
// output, so they drive the shell mock instead of taking its empty default.
func Test_ShellCmd_PositionalOrder_One(t *testing.T) {
	script := `
a = $"cmd"
print("a={a}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("OUT", "ERR", 0))
	assertOnlyOutput(t, stdOutBuffer, "a=OUT\n")
	assertNoErrors(t)
}

func Test_ShellCmd_PositionalOrder_Two(t *testing.T) {
	script := `
a, b = $"cmd"
print("a={a} b={b}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("OUT", "ERR", 0))
	assertOnlyOutput(t, stdOutBuffer, "a=OUT b=ERR\n")
	assertNoErrors(t)
}

func Test_ShellCmd_PositionalOrder_Three(t *testing.T) {
	// Exit code last. It's 0 here because a non-zero exit would propagate
	// before the print - see Test_ShellCmd_PositionalOrder_ThreeWithCatch.
	script := `
a, b, c = $"cmd"
print("a={a} b={b} c={c} {type_of(c)}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("OUT", "ERR", 0))
	assertOnlyOutput(t, stdOutBuffer, "a=OUT b=ERR c=0 int\n")
	assertNoErrors(t)
}

func Test_ShellCmd_PositionalOrder_ThreeWithCatch(t *testing.T) {
	// A `catch:` is the only way to observe a non-zero code: targets are
	// assigned before the exit check, then the error propagates.
	script := `
a, b, c = $"cmd" catch:
    print("a={a} b={b} c={c}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("OUT", "ERR", 3))
	assertOnlyOutput(t, stdOutBuffer, "a=OUT b=ERR c=3\n")
	assertNoErrors(t)
}

func Test_ShellCmd_NamedOrder_MatchesPositional(t *testing.T) {
	// The canonical prefix reads the same either way - this is the property
	// the (stdout, stderr, code) order buys us.
	script := `
stdout, stderr, code = $"cmd"
print("{stdout} {stderr} {code}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("OUT", "ERR", 0))
	assertOnlyOutput(t, stdOutBuffer, "OUT ERR 0\n")
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
	// Mixed naming falls back to positional, so the reserved name is bound by
	// slot, not by meaning: `stderr` gets stdout and `myvar` gets stderr. The
	// checker warns about exactly this shape (RAD40018).
	script := `
stderr, myvar = $"cmd"
print("stderr={stderr} myvar={myvar}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("OUT", "ERR", 0))

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "cmd",
		CaptureStdout: true,
		CaptureStderr: true,
		IsQuiet:       false,
		IsConfirm:     false,
	})
	assertOnlyOutput(t, stdOutBuffer, "stderr=OUT myvar=ERR\n")
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
	assertErrorContains(t, 1, "RAD20048", "Command exited with code 1")
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
out, err, code = confirm $"echo hi" catch:
    print("code={code} out=[{out}] err=[{err}]")
`
	decline := func(title, prompt string) (bool, error) { return false, nil }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(decline))

	// Declining must still leave capture targets defined (empty), just like a
	// command that ran and exited non-zero. Regression: `out` was previously
	// left undefined, blowing up the catch block with an undefined-variable error.
	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertOnlyOutput(t, stdOutBuffer, "code=1 out=[] err=[]\n")
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

// --- Interpolation quoting -------------------------------------------------
//
// The contract: literal text in a command is shell, interpolations are data.
// These assert the exact string handed to the shell, because that string is
// where the guarantee either holds or doesn't.

func Test_ShellCmd_Interp_QuotesApostrophe(t *testing.T) {
	script := "msg = \"it's fine\"\n$`git commit -m {msg}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `git commit -m 'it'\''s fine'`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_QuotesSpace(t *testing.T) {
	script := "name = \"My Notes.txt\"\n$`cat {name}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `cat 'My Notes.txt'`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_NeutralizesInjection(t *testing.T) {
	script := "payload = `a\"; echo INJECTED; echo \"b`\n$`echo {payload}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `echo 'a"; echo INJECTED; echo "b'`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_QuotesShellMetacharacters(t *testing.T) {
	script := "vals = [\"$HOME\", \"*.txt\", \"a`b\", \"x;y\", \"p|q\"]\nfor v in vals:\n    $`echo {v}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t,
		core.ShellInvocation{Command: `echo '$HOME'`},
		core.ShellInvocation{Command: `echo '*.txt'`},
		core.ShellInvocation{Command: "echo 'a`b'"},
		core.ShellInvocation{Command: `echo 'x;y'`},
		core.ShellInvocation{Command: `echo 'p|q'`},
	)
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_QuotesNewline(t *testing.T) {
	script := "msg = \"a\\nb\"\n$`echo {msg}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: "echo 'a\nb'"})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_QuotesEmptyString(t *testing.T) {
	script := "msg = \"\"\n$`echo {msg}`"
	setupAndRunCode(t, script, "--color=never")

	// Without quoting, an empty value would vanish rather than be an argument.
	assertShellInvoked(t, core.ShellInvocation{Command: `echo ''`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_LeavesSafeValuesBare(t *testing.T) {
	// The overwhelmingly common case must stay byte-identical to pre-quoting
	// Rad, or every ⚡️ echo and every snapshot gets noisier for no gain.
	script := "v = \"1.2.3\"\np = \"src/main.go\"\nn = 42\n$`git tag v{v} {p} {n}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `git tag v1.2.3 src/main.go 42`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_AdjacentSegmentsFormOneWord(t *testing.T) {
	// The shell concatenates adjacent quoted fragments, so a value glues onto
	// neighbouring literal text without the script quoting anything.
	script := "n = \"My Notes.txt\"\nv = \"1.2\"\n$`cp --flag={n} {v}-{v}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `cp --flag='My Notes.txt' 1.2-1.2`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_AppliesFormatSpecBeforeQuoting(t *testing.T) {
	script := "n = 7\n$`echo {n:03}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `echo 007`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_ListSplatsToOneWordPerElement(t *testing.T) {
	script := "files = [\"My Notes.txt\", \"a'b.txt\", \"c.txt\"]\n$`rm {files}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `rm 'My Notes.txt' 'a'\''b.txt' c.txt`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_EmptyListContributesNoWords(t *testing.T) {
	script := "flags = []\n$`rg {flags} -- needle`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `rg  -- needle`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_LiteralShellSyntaxSurvives(t *testing.T) {
	script := "f = \"a b.txt\"\n$`cat {f} | grep x > out.txt && echo done`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `cat 'a b.txt' | grep x > out.txt && echo done`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_RawStringIsNotInterpolated(t *testing.T) {
	script := "$r`echo {notavar}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `echo {notavar}`})
	assertNoErrors(t)
}

func Test_ShellCmd_Interp_NullIsRejected(t *testing.T) {
	script := "x = null\n$`echo {x}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20045", "Cannot interpolate a null", "'??'")
}

func Test_ShellCmd_Interp_MapIsRejected(t *testing.T) {
	script := "m = { \"a\": 1 }\n$`echo {m}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20045", "Cannot interpolate a map")
}

func Test_ShellCmd_Interp_ListInsideAWordIsRejected(t *testing.T) {
	script := "f = [\"a\", \"b\"]\n$`echo --files={f}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20045", "must stand alone as its own argument")
}

func Test_ShellCmd_Interp_ListWithFormatSpecIsRejected(t *testing.T) {
	script := "f = [\"a\", \"b\"]\n$`echo {f:5}`"
	setupAndRunCode(t, script, "--color=never")

	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20045", "format specifier")
}

// --- Argv form -------------------------------------------------------------

func Test_ShellCmd_Argv_ListLiteralBypassesShell(t *testing.T) {
	script := "msg = \"it's fine\"\n$[\"git\", \"commit\", \"-m\", msg]"
	setupAndRunCode(t, script, "--color=never")

	// No Command: the argv form never becomes shell text, so there is nothing
	// for a shell to reinterpret.
	assertShellInvoked(t, core.ShellInvocation{
		Argv: []string{"git", "commit", "-m", "it's fine"},
	})
	assertNoErrors(t)
}

func Test_ShellCmd_Argv_ListVariable(t *testing.T) {
	script := "cmd = [\"echo\", \"a b\"]\ncmd += [\"c\"]\n$cmd"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Argv: []string{"echo", "a b", "c"}})
	assertNoErrors(t)
}

func Test_ShellCmd_Argv_CoercesScalars(t *testing.T) {
	script := "$[\"seq\", 1, 2.5, true]"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Argv: []string{"seq", "1", "2.5", "true"}})
	assertNoErrors(t)
}

func Test_ShellCmd_Argv_CapturesLikeTheStringForm(t *testing.T) {
	// Positional, so it fills (stdout, stderr, code) - the same order the
	// string form uses. Capture binding is a property of the assignment, not
	// of how the command was spelled.
	script := "out, err = $[\"echo\", \"hi\"]"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Argv:          []string{"echo", "hi"},
		CaptureStdout: true,
		CaptureStderr: true,
	})
	assertNoErrors(t)
}

func Test_ShellCmd_Argv_EmptyListIsRejected(t *testing.T) {
	script := "cmd = []\n$cmd"
	setupAndRunCode(t, script, "--color=never")

	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20045", "no program to run")
}

func Test_ShellCmd_Argv_NestedListIsRejected(t *testing.T) {
	script := "$[\"echo\", [\"a\"]]"
	setupAndRunCode(t, script, "--color=never")

	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20045", "Concatenate it into the command list")
}

func Test_ShellCmd_Argv_NonStringNonListCommandIsRejected(t *testing.T) {
	script := "cmd = 5\n$cmd"
	setupAndRunCode(t, script, "--color=never")

	assertShellNotInvoked(t)
	assertErrorContains(t, 1, "RAD20045", "must be a string or a list of arguments")
}

// --- The raw form ----------------------------------------------------------

func Test_ShellCmd_StringVariableStaysVerbatim(t *testing.T) {
	// Interpolation happened when the string was built, so there is nothing
	// left for `$` to quote. This is the documented escape hatch, and changing
	// it would break every script that assembles a command by hand.
	script := "d = \"a b\"\ncmd = `echo {d}`\n$cmd"
	setupAndRunCode(t, script, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{Command: `echo a b`})
	assertNoErrors(t)
}

// A command that cannot be started is an ordinary mistake - a binary that isn't
// installed - not a Rad bug. Before this was fixed it panicked with a bare Go
// string, which isn't a RadPanic, so `catch:` couldn't see it and the user got
// RAD20042 "This is a bug in Rad" plus a Go stack trace. These two run the real
// executor because spawning is exactly the part the mock stands in for.
func Test_ShellCmd_Argv_SpawnFailureIsCatchable(t *testing.T) {
	script := `
code = $["rad_no_such_binary_zzz"] catch:
    print("caught {code}")
print("continued")
`
	setupAndRun(t, NewTestParams(script, "--color=never").RealShell())

	assertOutput(t, stdOutBuffer, "caught 127\ncontinued\n")
	assertNoErrors(t)
}

func Test_ShellCmd_Argv_SpawnFailureExitsWith127(t *testing.T) {
	script := `$["rad_no_such_binary_zzz"]`
	setupAndRun(t, NewTestParams(script, "--color=never").RealShell())

	// Same surface the string form gets from the shell: 127, and a stderr line
	// naming what couldn't be found.
	assertErrorContains(t, 1,
		"rad_no_such_binary_zzz: command not found",
		"Command exited with code 127",
	)
}

// --- Reading a result inline: `$`cmd`.stdout` and friends ------------------

// Capture follows the accessor, exactly as it follows the target names in a
// named assignment. The stream you didn't ask for still reaches the terminal.
func Test_ShellExpr_Stdout_CapturesStdoutOnly(t *testing.T) {
	setupAndRunCode(t, `x = $"cmd".stdout`, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "cmd",
		CaptureStdout: true,
		CaptureStderr: false,
	})
	assertNoErrors(t)
}

func Test_ShellExpr_Stderr_CapturesStderrOnly(t *testing.T) {
	setupAndRunCode(t, `x = $"cmd".stderr`, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "cmd",
		CaptureStdout: false,
		CaptureStderr: true,
	})
	assertNoErrors(t)
}

// Asking only about the outcome captures nothing, so both streams pass
// through - a `code` target behaves the same way.
func Test_ShellExpr_Code_CapturesNeither(t *testing.T) {
	setupAndRunCode(t, `x = $"cmd".code`, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "cmd",
		CaptureStdout: false,
		CaptureStderr: false,
	})
	assertNoErrors(t)
}

func Test_ShellExpr_Ok_CapturesNeither(t *testing.T) {
	setupAndRunCode(t, `x = $"cmd".ok`, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "cmd",
		CaptureStdout: false,
		CaptureStderr: false,
	})
	assertNoErrors(t)
}

func Test_ShellExpr_AccessorValues(t *testing.T) {
	script := `
print($"cmd".stdout.trim())
print($"cmd".code)
print(type_of($"cmd".code))
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("OUT\n", "ERR", 0))
	assertOnlyOutput(t, stdOutBuffer, "OUT\n0\nint\n")
	assertNoErrors(t)
}

// The output accessors raise on a non-zero exit: stdout the command never
// produced is not an answer.
func Test_ShellExpr_Stdout_PropagatesNonZero(t *testing.T) {
	script := `
x = $"cmd".stdout
print("unreached")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("", "", 3))
	assertErrorContains(t, 1, "RAD20048", "Command exited with code 3")
}

// .ok and .code never raise - using them IS the handling. grep exiting 1 is
// data, not failure, and `if not $`which x`.ok:` has to be writable.
func Test_ShellExpr_Ok_DoesNotPropagate(t *testing.T) {
	script := `
if not $"cmd".ok:
    print("failed, handled")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("", "", 3))
	assertOnlyOutput(t, stdOutBuffer, "failed, handled\n")
	assertNoErrors(t)
}

func Test_ShellExpr_Code_DoesNotPropagate(t *testing.T) {
	script := `
code = $"cmd".code
print("code={code}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("", "", 3))
	assertOnlyOutput(t, stdOutBuffer, "code=3\n")
	assertNoErrors(t)
}

// The regression test for the footgun this feature exists to remove. Before
// the grammar change `catch` bound to the command *string*, so the handler
// never ran and the failure killed the script.
func Test_ShellExpr_ComposesWithCatchOperator(t *testing.T) {
	script := `
x = $"cmd".stdout catch "fallback"
print("x={x}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("", "", 1))
	assertOnlyOutput(t, stdOutBuffer, "x=fallback\n")
	assertNoErrors(t)
}

func Test_ShellExpr_ComposesWithFallbackOperator(t *testing.T) {
	script := `
x = $"cmd".stdout ?? "fallback"
print("x={x}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("", "", 1))
	assertOnlyOutput(t, stdOutBuffer, "x=fallback\n")
	assertNoErrors(t)
}

func Test_ShellExpr_ComposesWithCatchBlock(t *testing.T) {
	script := `
x = $"cmd".stdout catch:
    x = "fallback"
print("x={x}")
`
	setupAndRun(t, NewTestParams(script, "--color=never").ShellOutput("", "", 1))
	assertOnlyOutput(t, stdOutBuffer, "x=fallback\n")
	assertNoErrors(t)
}

func Test_ShellExpr_QuietModifier(t *testing.T) {
	setupAndRunCode(t, `x = quiet $"cmd".stdout`, "--color=never")

	assertShellInvoked(t, core.ShellInvocation{
		Command:       "cmd",
		CaptureStdout: true,
		IsQuiet:       true,
	})
	assertNoErrors(t)
}

// Declining a confirm reports exit 1 without running, so `.ok` is false and
// nothing new had to be invented for the inline form.
func Test_ShellExpr_ConfirmDeclineIsNotOk(t *testing.T) {
	script := `
if not confirm $"rm -rf /".ok:
    print("declined")
`
	decline := func(title, prompt string) (bool, error) { return false, nil }
	setupAndRun(t, NewTestParams(script, "--color=never").ConfirmResponder(decline))

	assertConfirmCount(t, 1)
	assertShellNotInvoked(t)
	assertOnlyOutput(t, stdOutBuffer, "declined\n")
	assertNoErrors(t)
}
