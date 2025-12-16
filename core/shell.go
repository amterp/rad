package core

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
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

func (i *Interpreter) executeShellStmt(shellStmtNode *ts.Node) EvalResult {
	// Get left-hand variables for assignment
	leftNode := rl.GetChildren(shellStmtNode, rl.F_LEFT)
	leftNodes := rl.GetChildren(shellStmtNode, rl.F_LEFTS)
	leftNodes = append(leftNode, leftNodes...)

	if len(leftNodes) > 3 {
		i.errorf(shellStmtNode, "At most 3 assignments allowed with shell commands")
	}

	// Determine if using named assignment (all vars are code/stdout/stderr)
	isNamedAssignment := i.isNamedShellAssignment(leftNodes)

	// Determine capture mode based on assignment type
	var captureStdout, captureStderr bool
	if isNamedAssignment {
		captureStdout, captureStderr = i.namedCaptureMode(leftNodes)
	} else {
		captureStdout, captureStderr = i.positionalCaptureMode(len(leftNodes))
	}

	// Get catch block and shell command nodes
	catchBlockNode := rl.GetChild(shellStmtNode, rl.F_CATCH)
	shellCmdNode := rl.GetChild(shellStmtNode, rl.F_SHELL_CMD)

	// Helper to assign shell results to variables
	assignResults := func(result shellResult) {
		i.assignShellResults(shellCmdNode, leftNodes, result, isNamedAssignment)
	}

	return i.withCatch(catchBlockNode, func(rp *RadPanic) EvalResult {
		// Error occurred during shell execution (non-zero exit code)
		// The RadPanic contains the shell result via a special field
		result := rp.ShellResult

		// Assign variables to actual shell command results
		assignResults(*result)

		// Run catch block statements
		stmtNodes := rl.GetChildren(catchBlockNode, rl.F_STMT)
		res := i.runBlock(stmtNodes)
		if res.Ctrl != CtrlNormal {
			return res // Propagate control flow (return/break/continue)
		}
		return VoidNormal
	}, func() EvalResult {
		// Normal execution - run shell command
		result := i.executeShellCmd(shellCmdNode, captureStdout, captureStderr)

		// Assign variables to shell command results
		assignResults(result)

		// If exit code != 0, propagate error
		if result.exitCode != 0 {
			// Create a RadPanic with the shell result embedded
			err := NewErrorStrf("Command exited with code %d", result.exitCode).SetNode(shellCmdNode)
			rp := &RadPanic{
				ErrV:        newRadValue(i, shellCmdNode, err),
				ShellResult: &result,
			}
			panic(rp)
		}

		return VoidNormal
	})
}

func (i *Interpreter) positionalCaptureMode(numVars int) (captureStdout bool, captureStderr bool) {
	// Positional: 0 vars = nothing, 1 var = code only, 2 vars = code+stdout, 3 vars = code+stdout+stderr
	captureStdout = numVars >= 2
	captureStderr = numVars >= 3
	return
}

func (i *Interpreter) namedCaptureMode(leftNodes []ts.Node) (captureStdout bool, captureStderr bool) {
	// Named: capture only what's requested
	for _, node := range leftNodes {
		rootNode := rl.GetChild(&node, rl.F_ROOT)
		varName := i.GetSrcForNode(rootNode)
		switch varName {
		case "stdout":
			captureStdout = true
		case "stderr":
			captureStderr = true
		}
	}
	return
}

func (i *Interpreter) executeShellCmd(shellCmdNode *ts.Node, captureStdout, captureStderr bool) shellResult {
	// Check for modifiers by inspecting all modifier nodes
	modifierNodes := rl.GetChildren(shellCmdNode, rl.F_MODIFIER)
	var isQuiet, isConfirm bool
	// todo they should be using different field names, really
	for _, modNode := range modifierNodes {
		modText := i.GetSrcForNode(&modNode)
		switch modText {
		case "quiet":
			isQuiet = true
		case "confirm":
			isConfirm = true
		}
	}

	// evaluate the command string
	cmdNode := rl.GetChild(shellCmdNode, rl.F_COMMAND)
	cmdStr := i.eval(cmdNode).Val.
		RequireType(i, cmdNode, "Shell commands must be strings", rl.RadStrT).
		RequireStr(i, shellCmdNode)

	// Create invocation and execute via RShell
	invocation := ShellInvocation{
		Command:       cmdStr.Plain(),
		CaptureStdout: captureStdout,
		CaptureStderr: captureStderr,
		IsQuiet:       isQuiet,
		IsConfirm:     isConfirm,
	}

	stdout, stderr, exitCode := RShell(invocation)

	// Build result from RShell output
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
	// Handle confirmation prompt if needed
	if FlagConfirmShellCommands.Value || invocation.IsConfirm {
		ok, err := InputConfirm(invocation.Command, "Run above command? [y/n] > ")
		if err != nil {
			// Can't use i.errorf here, so just panic
			panic(fmt.Sprintf("Error confirming shell command: %v", err))
		}
		if !ok {
			return "", "", 1
		}
	}

	// Build the command
	cmd := resolveCmdSimple(invocation.Command)
	var stdoutBuf, stderrBuf bytes.Buffer

	// Set up output destinations based on capture mode
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

	// Print command if not quiet
	if !invocation.IsQuiet {
		RP.RadStderrf("⚡️ %s\n", invocation.Command)
	}

	// Run the command
	err := cmd.Run()

	// Handle exit codes and errors
	exitCode := 0
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode = exitErr.ExitCode()
		} else {
			// Non-exit error (e.g., command not found)
			panic(fmt.Sprintf("Failed to run command: %v\nStderr: %s\n", err, stderrBuf.String()))
		}
	}

	// Only return captured output
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

// resolveCmdSimple builds a shell command without needing interpreter context
func resolveCmdSimple(cmdStr string) *exec.Cmd {
	// check SHELL first - most accurate reflection of the environment
	if shell := os.Getenv("SHELL"); shell != "" {
		return buildCmd(shell, cmdStr)
	}

	// last resort for Unix-like systems
	if _, err := exec.LookPath("/bin/sh"); err == nil {
		return buildCmd("/bin/sh", cmdStr)
	}

	// this is also where we could detect and allow windows commands, if we wanted.

	panic("Cannot run shell cmd as no shell found. Please set the SHELL environment variable.")
}

func resolveCmd(i *Interpreter, shellNode *ts.Node, cmdStr string) *exec.Cmd {
	// todo potentially want to somehow inject a flag which makes pipe commands propagate errors

	// check SHELL first - most accurate reflection of the environment
	if shell := os.Getenv("SHELL"); shell != "" {
		return buildCmd(shell, cmdStr)
	}

	// last resort for Unix-like systems
	if _, err := exec.LookPath("/bin/sh"); err == nil {
		return buildCmd("/bin/sh", cmdStr)
	}

	// this is also where we could detect and allow windows commands, if we wanted.

	i.errorf(shellNode, "Cannot run shell cmd as no shell found. Please set the SHELL environment variable.")
	panic(UNREACHABLE)
}

func buildCmd(shellStr string, cmdStr string) *exec.Cmd {
	cmd := exec.Command(shellStr, "-c", cmdStr)
	// if we don't unwrap to Stdin *file*, the command never ends? It's a bit weird, should understand better.
	cmd.Stdin = RIo.StdIn.Unwrap()

	// cmd.Stderr = RIo.StdErr // todo ? this seems to conflict with the pipes later
	// cmd.Stdout = RIo.StdOut

	return cmd
}

// isNamedShellAssignment checks if ALL variables are named exactly "code", "stdout", or "stderr"
// If so, assignment is by name (order independent). Otherwise, positional.
func (i *Interpreter) isNamedShellAssignment(leftNodes []ts.Node) bool {
	if len(leftNodes) == 0 {
		return false
	}

	for _, node := range leftNodes {
		// Get the root identifier from the var_path
		rootNode := rl.GetChild(&node, rl.F_ROOT)
		if rootNode == nil {
			return false
		}

		varName := i.GetSrcForNode(rootNode)
		if varName != "code" && varName != "stdout" && varName != "stderr" {
			return false
		}
	}

	return true
}

// assignShellResults assigns shell command results to variables
// Uses named or positional assignment depending on isNamedAssignment flag
func (i *Interpreter) assignShellResults(
	shellCmdNode *ts.Node,
	leftNodes []ts.Node,
	result shellResult,
	isNamedAssignment bool,
) {
	if isNamedAssignment {
		// Named assignment: match by variable name regardless of order
		for _, node := range leftNodes {
			rootNode := rl.GetChild(&node, rl.F_ROOT)
			varName := i.GetSrcForNode(rootNode)

			switch varName {
			case "code":
				i.doVarPathAssign(&node, newRadValue(i, shellCmdNode, int64(result.exitCode)), false)
			case "stdout":
				if result.stdout != nil {
					i.doVarPathAssign(&node, newRadValue(i, shellCmdNode, *result.stdout), false)
				}
			case "stderr":
				if result.stderr != nil {
					i.doVarPathAssign(&node, newRadValue(i, shellCmdNode, *result.stderr), false)
				}
			}
		}
	} else {
		// Positional assignment: code, stdout, stderr in order
		if len(leftNodes) >= 1 {
			i.doVarPathAssign(&leftNodes[0], newRadValue(i, shellCmdNode, int64(result.exitCode)), false)
		}
		if len(leftNodes) >= 2 && result.stdout != nil {
			i.doVarPathAssign(&leftNodes[1], newRadValue(i, shellCmdNode, *result.stdout), false)
		}
		if len(leftNodes) >= 3 && result.stderr != nil {
			i.doVarPathAssign(&leftNodes[2], newRadValue(i, shellCmdNode, *result.stderr), false)
		}
	}
}
