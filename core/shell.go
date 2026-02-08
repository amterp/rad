package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

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
	IsConfirm     bool
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
		result := rp.ShellResult
		assignResults(*result)

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

	stdout, stderr, exitCode := RShell(invocation)

	result := shellResult{
		exitCode: exitCode,
	}

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
	if FlagConfirmShellCommands.Value || invocation.IsConfirm {
		ok, err := InputConfirm(invocation.Command, "Run above command? [y/n] > ")
		if err != nil {
			panic(fmt.Sprintf("Error confirming shell command: %v", err))
		}
		if !ok {
			return "", "", 1
		}
	}

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

func resolveCmdSimple(cmdStr string) *exec.Cmd {
	if shell := os.Getenv("SHELL"); shell != "" {
		return buildCmd(shell, cmdStr)
	}

	if _, err := exec.LookPath("/bin/sh"); err == nil {
		return buildCmd("/bin/sh", cmdStr)
	}

	panic("Cannot run shell cmd as no shell found. Please set the SHELL environment variable.")
}

func resolveCmd(i *Interpreter, shellNode rl.Node, cmdStr string) *exec.Cmd {
	if shell := os.Getenv("SHELL"); shell != "" {
		return buildCmd(shell, cmdStr)
	}

	if _, err := exec.LookPath("/bin/sh"); err == nil {
		return buildCmd("/bin/sh", cmdStr)
	}

	i.emitError(rl.ErrGenericRuntime, shellNode, "Cannot run shell cmd as no shell found. Please set the SHELL environment variable")
	panic(UNREACHABLE)
}

func buildCmd(shellStr string, cmdStr string) *exec.Cmd {
	cmd := exec.Command(shellStr, "-c", cmdStr)
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
