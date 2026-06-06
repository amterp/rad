package core

import (
	"bytes"
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

// ShellInvocation captures the details of a shell command invocation
type ShellInvocation struct {
	Command       string
	CaptureStdout bool
	CaptureStderr bool
	IsQuiet       bool
	// IsConfirm is informational: the confirmation prompt is handled in
	// executeShellCmd before the executor runs, so executors don't act on it.
	// Kept for callers/observability (e.g. test assertions).
	IsConfirm bool
}

// ShellExecutor is the function type for executing shell commands
// Returns: (stdout, stderr, exitCode) - only returns captured output based on invocation.Capture* fields
type ShellExecutor func(invocation ShellInvocation) (string, string, int)

func (i *Interpreter) executeShellStmt(shell *rl.Shell) EvalResult {
	targets := shell.Targets

	if len(targets) > 3 {
		i.emitError(rl.ErrInvalidSyntax, shell, "At most 3 assignments allowed with shell commands")
	}

	// Determine if using named assignment (all vars are code/stdout/stderr)
	isNamedAssignment := isNamedShellAssignment(targets)

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

func positionalCaptureMode(numVars int) (captureStdout bool, captureStderr bool) {
	captureStdout = numVars >= 2
	captureStderr = numVars >= 3
	return
}

func namedCaptureMode(targets []rl.Node) (captureStdout bool, captureStderr bool) {
	for _, target := range targets {
		name := extractRootName(target)
		switch name {
		case "stdout":
			captureStdout = true
		case "stderr":
			captureStderr = true
		}
	}
	return
}

func (i *Interpreter) executeShellCmd(shell *rl.Shell) shellResult {
	cmdStr := i.eval(shell.Cmd).Val.
		RequireType(i, shell.Cmd, "Shell commands must be strings", rl.RadStrT).
		RequireStr(i, shell)

	captureStdout, captureStderr := namedCaptureMode(shell.Targets)
	if !isNamedShellAssignment(shell.Targets) {
		captureStdout, captureStderr = positionalCaptureMode(len(shell.Targets))
	}

	invocation := ShellInvocation{
		Command:       cmdStr.Plain(),
		CaptureStdout: captureStdout,
		CaptureStderr: captureStderr,
		IsQuiet:       shell.IsQuiet,
		IsConfirm:     shell.IsConfirm,
	}

	if FlagConfirmShellCommands.Value || shell.IsConfirm {
		ok, err := RConfirm(invocation.Command, "Run above command? [Y/n] > ")
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

	stdout, stderr, exitCode := RShell(invocation)
	return newShellResult(exitCode, stdout, stderr, captureStdout, captureStderr)
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
func realShellExecutor(invocation ShellInvocation) (string, string, int) {
	cmd := resolveCmdSimple(invocation.Command)
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
		RP.RadStderrf("⚡️ %s\n", invocation.Command)
	}

	err := cmd.Run()

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

// extractRootName gets the identifier name from a VarPath or Identifier AST node.
func extractRootName(node rl.Node) string {
	switch n := node.(type) {
	case *rl.VarPath:
		if ident, ok := n.Root.(*rl.Identifier); ok {
			return ident.Name
		}
	case *rl.Identifier:
		return n.Name
	}
	return ""
}

// isNamedShellAssignment checks if ALL variables are named exactly "code", "stdout", or "stderr"
func isNamedShellAssignment(targets []rl.Node) bool {
	if len(targets) == 0 {
		return false
	}

	for _, target := range targets {
		name := extractRootName(target)
		if name != "code" && name != "stdout" && name != "stderr" {
			return false
		}
	}

	return true
}

// assignShellResults assigns shell command results to variables
func (i *Interpreter) assignShellResults(
	shell *rl.Shell,
	targets []rl.Node,
	result shellResult,
	isNamedAssignment bool,
) {
	if isNamedAssignment {
		for _, target := range targets {
			name := extractRootName(target)
			switch name {
			case "code":
				i.doVarPathAssign(target, newRadValue(i, shell, int64(result.exitCode)), false)
			case "stdout":
				if result.stdout != nil {
					i.doVarPathAssign(target, newRadValue(i, shell, *result.stdout), false)
				}
			case "stderr":
				if result.stderr != nil {
					i.doVarPathAssign(target, newRadValue(i, shell, *result.stderr), false)
				}
			}
		}
	} else {
		if len(targets) >= 1 {
			i.doVarPathAssign(targets[0], newRadValue(i, shell, int64(result.exitCode)), false)
		}
		if len(targets) >= 2 && result.stdout != nil {
			i.doVarPathAssign(targets[1], newRadValue(i, shell, *result.stdout), false)
		}
		if len(targets) >= 3 && result.stderr != nil {
			i.doVarPathAssign(targets[2], newRadValue(i, shell, *result.stderr), false)
		}
	}
}
