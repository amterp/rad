package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/amterp/rad/rts/rl"
)

// todo
//  - implement mocking shell responses, like with json requests
//  - tests!
//  - improve error output, especially when stderr is not captured, because that prints first then, before Rad handles it
//  - silent keyword to suppress output?

type shellResult struct {
	exitCode int
	stdout   *string
	stderr   *string
}

// ShellInvocation captures the details of a shell command invocation.
//
// Exactly one of Command or Argv is set, and they mean different things.
// Command is a shell *program*: it goes to `<shell> -c <string>` and the shell
// parses pipes, redirects and expansions out of it. Argv is a literal argument
// vector: it is exec'd directly with no shell in the path, so nothing in it can
// be reinterpreted as syntax.
type ShellInvocation struct {
	Command       string
	Argv          []string
	CaptureStdout bool
	CaptureStderr bool
	IsQuiet       bool
	// IsConfirm is informational: the confirmation prompt is handled in
	// executeShellCmd before the executor runs, so executors don't act on it.
	// Kept for callers/observability (e.g. test assertions).
	IsConfirm bool
}

// IsArgv reports whether this invocation bypasses the shell.
func (s ShellInvocation) IsArgv() bool {
	return s.Argv != nil
}

// Display renders the invocation as a string a user could paste into their own
// shell. For the shell-string form that's the command verbatim; for the argv
// form we re-quote each element, since the argv never was shell text and has no
// faithful flat rendering. Used for the ⚡️ echo and the confirm prompt, never
// for execution.
func (s ShellInvocation) Display() string {
	if !s.IsArgv() {
		return s.Command
	}
	quoted := make([]string, 0, len(s.Argv))
	for _, arg := range s.Argv {
		quoted = append(quoted, shellQuoteIfNeeded(arg))
	}
	return strings.Join(quoted, " ")
}

// ShellExecutor is the function type for executing shell commands.
// Returns: (stdout, stderr, exitCode) - only returns captured output based on invocation.Capture* fields.
// The ctx is the interpreter's signal-cancellation context; the executor should wake up promptly when
// it's canceled (the subprocess shares Rad's process group, so it will have received the signal too).
type ShellExecutor func(ctx context.Context, invocation ShellInvocation) (string, string, int)

func (i *Interpreter) executeShellStmt(shell *rl.Shell) EvalResult {
	targets := shell.Targets

	if len(targets) > 3 {
		i.emitError(rl.ErrInvalidSyntax, shell, "At most 3 assignments allowed with shell commands")
	}

	// Determine if using named assignment (all vars are code/stdout/stderr)
	isNamedAssignment := rl.IsNamedShellAssignment(targets)

	// Helper to assign shell results to variables
	assignResults := func(result shellResult) {
		i.assignShellResults(shell, targets, result, isNamedAssignment)
	}

	return i.withCatch(shell.Catch, func(rp *RadPanic) EvalResult {
		if rp.ShellResult != nil {
			assignResults(*rp.ShellResult)
		} else if len(targets) > 0 {
			// The panic came from the command expression itself (not a shell exit code),
			// so there's no shell result. Assign the error to the first target and null to the rest,
			// matching how assignment catch handlers work.
			i.doVarPathAssign(targets[0], rp.ErrV, false)
			for j := 1; j < len(targets); j++ {
				i.doVarPathAssign(targets[j], RAD_NULL_VAL, false)
			}
		}

		res := i.runBlock(shell.Catch.Stmts)
		if res.Ctrl != CtrlNormal {
			return res
		}
		return VoidNormal
	}, func() EvalResult {
		// Normal execution
		result := i.executeShellCmd(shell)

		assignResults(result)

		if result.exitCode != 0 {
			err := NewErrorStrf("Command exited with code %d", result.exitCode).SetSpan(nodeSpanPtr(shell))
			rp := &RadPanic{
				ErrV:        newRadValue(i, shell, err),
				ShellResult: &result,
			}
			panic(rp)
		}

		return VoidNormal
	})
}

func (i *Interpreter) executeShellCmd(shell *rl.Shell) shellResult {
	command, argv := i.evalShellCommand(shell)

	captureStdout, captureStderr := rl.ShellCaptures(shell.Targets)

	invocation := ShellInvocation{
		Command:       command,
		Argv:          argv,
		CaptureStdout: captureStdout,
		CaptureStderr: captureStderr,
		IsQuiet:       shell.IsQuiet,
		IsConfirm:     shell.IsConfirm,
	}

	if FlagConfirmShellCommands.Value || shell.IsConfirm {
		ok, err := RConfirm(invocation.Display(), "Run above command? [Y/n] > ")
		if err != nil {
			// User aborted the prompt (Ctrl-C / Esc). Surface a catchable
			// user-input error, consistent with confirm()/pick()/input(),
			// rather than crashing as an internal bug.
			errVal := newRadValue(i, shell, NewErrorStrf("Shell command aborted: %v", err).SetCode(rl.ErrUserInput))
			i.NewRadPanic(shell, errVal).Panic()
		}
		if !ok {
			// User declined ("n"): don't run the command, but still surface
			// exit code 1 (a catchable "Command exited with code 1"). Populate
			// captures with empty output so capture targets stay defined, just
			// like a command that actually ran and exited non-zero would.
			return newShellResult(1, "", "", captureStdout, captureStderr)
		}
	}

	stdout, stderr, exitCode := RShell(i.signals.Ctx(), invocation)
	return newShellResult(exitCode, stdout, stderr, captureStdout, captureStderr)
}

// evalShellCommand evaluates a shell command expression into the pieces of a
// ShellInvocation. Exactly one of the two returns is populated.
//
// Which form you get is decided by what was written, because that's what says
// who is responsible for the shell syntax:
//
//	$`literal {x}`  - a command literal; the text is shell, {x} is data
//	$list           - an argument vector, exec'd with no shell involved
//	$str_expr       - a command string the script assembled itself, used verbatim
func (i *Interpreter) evalShellCommand(shell *rl.Shell) (command string, argv []string) {
	if lit, ok := shell.Cmd.(*rl.LitString); ok {
		return i.evalShellCmdString(lit), nil
	}

	val := i.eval(shell.Cmd).Val
	switch val.Type() {
	case rl.RadStrT:
		return val.RequireStr(i, shell).Plain(), nil
	case rl.RadListT:
		return "", i.shellArgvFromList(shell, val.RequireList(i, shell.Cmd))
	default:
		i.emitErrorf(rl.ErrShellCmdValue, shell.Cmd,
			"Shell commands must be a string or a list of arguments, got %s", TypeAsString(val))
		panic(UNREACHABLE)
	}
}

// evalShellCmdString evaluates a command literal, quoting every interpolated
// value so it reaches the program as written.
//
// The rule is that the literal text you typed is shell - pipes, redirects and
// all - while interpolations are data. A scalar becomes exactly one shell word;
// a list becomes one word per element. Because the shell concatenates adjacent
// quoted fragments, `--flag={x}` and `{a}-{b}` are still single words without
// the script quoting anything itself.
func (i *Interpreter) evalShellCmdString(n *rl.LitString) string {
	if n.Simple {
		return n.Value
	}

	var sb strings.Builder
	for idx, seg := range n.Segments {
		if seg.IsLiteral {
			sb.WriteString(seg.Text)
			continue
		}
		sb.WriteString(i.shellInterpolation(n, idx))
	}
	return sb.String()
}

// shellInterpolation renders one interpolation of a command literal as shell
// text: quoted so the shell reads it back as the exact value it started as.
func (i *Interpreter) shellInterpolation(n *rl.LitString, idx int) string {
	seg := n.Segments[idx]
	val := i.eval(seg.Expr).Val
	val.RequireNonVoid(i, seg.Expr)

	if val.Type() == rl.RadListT {
		return i.shellListInterpolation(n, idx, val.RequireList(i, seg.Expr))
	}

	if seg.Format != nil {
		return shellQuoteIfNeeded(i.formatInterpolation(seg, val).Plain())
	}
	return shellQuoteIfNeeded(i.shellScalarString(seg.Expr, val))
}

// shellListInterpolation splats a list into one quoted word per element.
//
// We require the interpolation to stand alone as its own word. `{files}` is
// N arguments, but `--file={files}` has no meaning we could pick that wouldn't
// surprise someone - repeating the prefix and gluing it to the first element
// are both defensible, so we make the script say which it wants.
func (i *Interpreter) shellListInterpolation(n *rl.LitString, idx int, list *RadList) string {
	seg := n.Segments[idx]

	if seg.Format != nil {
		i.emitError(rl.ErrShellCmdValue, seg.Expr,
			"Cannot apply a format specifier to a list in a shell command, "+
				"because the list becomes several arguments rather than one string")
		panic(UNREACHABLE)
	}

	if !segmentStandsAlone(n, idx) {
		i.emitError(rl.ErrShellCmdValue, seg.Expr,
			"A list in a shell command must stand alone as its own argument, "+
				"because it expands to one argument per element. "+
				"Surround it with spaces, or join it into a single string first")
		panic(UNREACHABLE)
	}

	parts := make([]string, 0, list.LenInt())
	for _, elem := range list.Values {
		parts = append(parts, shellQuoteIfNeeded(i.shellScalarString(seg.Expr, elem)))
	}
	return strings.Join(parts, " ")
}

// segmentStandsAlone reports whether the segment at idx is delimited by
// whitespace (or the ends of the command) on both sides, i.e. whether it forms
// a whole shell word by itself.
func segmentStandsAlone(n *rl.LitString, idx int) bool {
	endsInSpace := func(prev int) bool {
		if prev < 0 {
			return true
		}
		seg := n.Segments[prev]
		return seg.IsLiteral && seg.Text != "" && isShellSpace(seg.Text[len(seg.Text)-1])
	}
	startsWithSpace := func(next int) bool {
		if next >= len(n.Segments) {
			return true
		}
		seg := n.Segments[next]
		return seg.IsLiteral && seg.Text != "" && isShellSpace(seg.Text[0])
	}
	return endsInSpace(idx-1) && startsWithSpace(idx+1)
}

func isShellSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

// shellScalarString renders a single value as the text of one shell word,
// before quoting. Non-scalars are rejected for the same reason as in the argv
// form: there is no rendering of a map or a null that isn't a silent bug.
func (i *Interpreter) shellScalarString(node rl.Node, val RadValue) string {
	switch val.Type() {
	case rl.RadStrT:
		return val.RequireStr(i, node).Plain()
	case rl.RadIntT, rl.RadFloatT, rl.RadBoolT:
		return ToPrintable(val)
	default:
		i.emitErrorf(rl.ErrShellCmdValue, node,
			"Cannot interpolate a %s into a shell command. %s",
			TypeAsString(val), shellArgFixHint(val.Type()))
		panic(UNREACHABLE)
	}
}

// shellArgvFromList converts a list command into an argument vector. Every
// element becomes exactly one argument, so a list is the safe way to build a
// command out of values that might contain spaces or shell metacharacters.
func (i *Interpreter) shellArgvFromList(shell *rl.Shell, list *RadList) []string {
	if list.IsEmpty() {
		i.emitError(rl.ErrShellCmdValue, shell.Cmd,
			"Shell command list is empty, so there is no program to run")
		panic(UNREACHABLE)
	}

	argv := make([]string, 0, list.LenInt())
	for idx, elem := range list.Values {
		argv = append(argv, i.shellArgString(shell, elem, idx))
	}
	return argv
}

// shellArgString renders one list element as a single argument. Scalars have an
// obvious one-argument spelling; anything else does not, and guessing one is how
// a `null` silently becomes the four-character word "null" in a command.
func (i *Interpreter) shellArgString(shell *rl.Shell, val RadValue, idx int) string {
	switch val.Type() {
	case rl.RadStrT:
		return val.RequireStr(i, shell.Cmd).Plain()
	case rl.RadIntT, rl.RadFloatT, rl.RadBoolT:
		return ToPrintable(val)
	default:
		i.emitErrorf(rl.ErrShellCmdValue, shell.Cmd,
			"Shell command argument %d is a %s, which has no single-argument form. %s",
			idx, TypeAsString(val), shellArgFixHint(val.Type()))
		panic(UNREACHABLE)
	}
}

// shellArgFixHint suggests a remedy for a value that can't be one argument.
func shellArgFixHint(t rl.RadType) string {
	switch t {
	case rl.RadNullT:
		return "Supply a fallback with '??', or drop the argument when it's null."
	case rl.RadListT:
		return "Concatenate it into the command list instead of nesting it."
	default:
		return "Convert it to a string first."
	}
}

// newShellResult assembles a shellResult, attaching captured stdout/stderr only
// when the corresponding capture was requested.
func newShellResult(exitCode int, stdout, stderr string, captureStdout, captureStderr bool) shellResult {
	result := shellResult{exitCode: exitCode}
	if captureStdout {
		result.stdout = &stdout
	}
	if captureStderr {
		result.stderr = &stderr
	}
	return result
}

// realShellExecutor is the production implementation of shell command execution
// warning: as of writing, this is *not* covered in tests
func realShellExecutor(ctx context.Context, invocation ShellInvocation) (string, string, int) {
	cmd := resolveCmd(invocation)
	var stdoutBuf, stderrBuf bytes.Buffer

	if invocation.CaptureStdout {
		cmd.Stdout = &stdoutBuf
	} else {
		cmd.Stdout = RIo.StdOut
	}

	if invocation.CaptureStderr {
		cmd.Stderr = &stderrBuf
	} else {
		cmd.Stderr = RIo.StdErr
	}

	if !invocation.IsQuiet {
		RP.RadStderrf("⚡️ %s\n", invocation.Display())
	}

	// Start+Wait with a select on ctx, so a signal arriving during the
	// subprocess wakes us promptly. Because we use a shared process group
	// (today's default), the subprocess will have received the same signal
	// and is expected to terminate on its own. We still wait for its actual
	// exit code so we can report it accurately.
	if err := cmd.Start(); err != nil {
		panic(fmt.Sprintf("Failed to start command: %v\n", err))
	}

	waitErr := make(chan error, 1)
	go func() {
		waitErr <- cmd.Wait()
	}()

	var err error
	select {
	case err = <-waitErr:
		// Subprocess finished normally (or with an error). Continue below.
	case <-ctx.Done():
		// A signal fired. The subprocess shares our process group and will
		// have received the signal too. Wait for it to exit so we get the
		// real exit code rather than guessing - if it ignores the signal
		// (rare), a second Ctrl+C force-exits Rad via the SignalManager.
		err = <-waitErr
	}

	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			panic(fmt.Sprintf("Failed to run command: %v\nStderr: %s\n", err, stderrBuf.String()))
		}
	}

	stdout := ""
	stderr := ""
	if invocation.CaptureStdout {
		stdout = stdoutBuf.String()
	}
	if invocation.CaptureStderr {
		stderr = stderrBuf.String()
	}

	return stdout, stderr, exitCode
}

// resolveCmd prepares an *exec.Cmd for either invocation form. Panics with a
// user-facing message if the shell-string form can't find a shell.
//
// The argv form deliberately does not resolve a shell at all - that's the whole
// point of it, and it's why this form works on platforms where we have no
// usable shell story.
func resolveCmd(invocation ShellInvocation) *exec.Cmd {
	if invocation.IsArgv() {
		cmd := exec.Command(invocation.Argv[0], invocation.Argv[1:]...)
		cmd.Stdin = RIo.StdIn.Unwrap()
		return cmd
	}
	return resolveCmdSimple(invocation.Command)
}

// resolveCmdSimple resolves the shell to use for the given command string and
// returns a prepared *exec.Cmd. Panics with a user-facing message if no shell
// can be found.
func resolveCmdSimple(cmdStr string) *exec.Cmd {
	path, flag, err := resolveShell(os.Getenv, exec.LookPath, IsWindows())
	if err != nil {
		panic(err.Error())
	}
	return buildCmd(path, flag, cmdStr)
}

// resolveShell picks a shell to use for executing a command string. It is a
// pure function: all platform/env dependencies are passed in so it is testable
// without mutating global state.
//
// This picks a shell but not a quoting scheme, and interpolated values are
// quoted as POSIX sh regardless (see shellQuoteIfNeeded). So a command carrying
// interpolations is wrong under any shell this can return that isn't POSIX:
// cmd.exe and PowerShell on Windows, and csh/tcsh or xonsh anywhere, if $SHELL
// names one. Fixing that means choosing the quoting from what we resolve here,
// or refusing the string form when we can't quote for it. RED-9 has the detail.
//
// Resolution order:
//  1. SHELL env var if set, but on Windows only if it actually resolves to an
//     executable. Git Bash / MSYS2 / Cygwin set SHELL to a Unix-style path
//     (e.g. /usr/bin/bash) that native Win32 exec can't find, so on Windows
//     we fall through to the candidate chain in that case rather than crash.
//  2. Windows: pwsh.exe -> powershell.exe -> cmd.exe
//  3. Other:   /bin/sh
//
// Returns the resolved shell path and the flag to use for command-string
// invocation (e.g. "-c" for POSIX shells and PowerShell, "/c" for cmd.exe).
func resolveShell(
	getEnv func(string) string,
	lookPath func(string) (string, error),
	isWindows bool,
) (path, flag string, err error) {
	if shell := strings.TrimSpace(getEnv("SHELL")); shell != "" {
		if !isWindows {
			return shell, shellExecFlag(shell), nil
		}
		// On Windows, only honor SHELL if it actually resolves - otherwise
		// fall through. This rescues the common Git Bash case where SHELL is
		// set to /usr/bin/bash but the native Win32 binary can't see it.
		if resolved, lookErr := lookPath(shell); lookErr == nil {
			return resolved, shellExecFlag(resolved), nil
		}
	}

	var candidates []string
	if isWindows {
		candidates = []string{"pwsh.exe", "powershell.exe", "cmd.exe"}
	} else {
		candidates = []string{"/bin/sh"}
	}

	for _, c := range candidates {
		if resolved, lookErr := lookPath(c); lookErr == nil {
			return resolved, shellExecFlag(resolved), nil
		}
	}

	return "", "", errors.New("Cannot run shell cmd as no shell found. Please set the SHELL environment variable")
}

// shellExecFlag returns the flag a given shell expects for invoking a command
// string. Defaults to "-c" (POSIX shells, bash/zsh, and PowerShell which
// accepts "-c" as a short form of "-Command"). Only cmd.exe needs "/c".
//
// We don't use filepath.Base because its separator handling is GOOS-specific
// (only "/" on Unix), which would mis-handle Windows-style paths that may
// arrive via env vars or mixed environments.
func shellExecFlag(shellPath string) string {
	if i := strings.LastIndexAny(shellPath, `/\`); i >= 0 {
		shellPath = shellPath[i+1:]
	}
	base := strings.TrimSuffix(strings.ToLower(shellPath), ".exe")
	if base == "cmd" {
		return "/c"
	}
	return "-c"
}

func buildCmd(shellStr string, flag string, cmdStr string) *exec.Cmd {
	cmd := exec.Command(shellStr, flag, cmdStr)
	cmd.Stdin = RIo.StdIn.Unwrap()
	return cmd
}

// assignShellResults binds each target to the stream it asked for. Which
// stream that is - by name or by position - is decided by rl.ShellStreamForTarget,
// shared with the checker so the two can't drift.
//
// A stream that wasn't captured has a nil pointer here and leaves its target
// unassigned. That only happens in named mode (positional always captures every
// stream it has a target for), where e.g. `code = $cmd` captures nothing.
func (i *Interpreter) assignShellResults(
	shell *rl.Shell,
	targets []rl.Node,
	result shellResult,
	isNamedAssignment bool,
) {
	for idx, target := range targets {
		switch rl.ShellStreamForTarget(targets, idx, isNamedAssignment) {
		case rl.ShellCode:
			i.doVarPathAssign(target, newRadValue(i, shell, int64(result.exitCode)), false)
		case rl.ShellStdout:
			if result.stdout != nil {
				i.doVarPathAssign(target, newRadValue(i, shell, *result.stdout), false)
			}
		case rl.ShellStderr:
			if result.stderr != nil {
				i.doVarPathAssign(target, newRadValue(i, shell, *result.stderr), false)
			}
		}
	}
}
