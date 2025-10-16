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

type shellCaptureMode struct {
	captureStdout bool
	captureStderr bool
}

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
	var captureMode shellCaptureMode
	if isNamedAssignment {
		captureMode = i.namedCaptureMode(leftNodes)
	} else {
		captureMode = i.positionalCaptureMode(len(leftNodes))
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
		result := i.executeShellCmd(shellCmdNode, captureMode)

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

func (i *Interpreter) positionalCaptureMode(numVars int) shellCaptureMode {
	// Positional: 0 vars = nothing, 1 var = code only, 2 vars = code+stdout, 3 vars = code+stdout+stderr
	return shellCaptureMode{
		captureStdout: numVars >= 2,
		captureStderr: numVars >= 3,
	}
}

func (i *Interpreter) namedCaptureMode(leftNodes []ts.Node) shellCaptureMode {
	// Named: capture only what's requested
	mode := shellCaptureMode{}
	for _, node := range leftNodes {
		rootNode := rl.GetChild(&node, rl.F_ROOT)
		varName := i.GetSrcForNode(rootNode)
		switch varName {
		case "stdout":
			mode.captureStdout = true
		case "stderr":
			mode.captureStderr = true
		}
	}
	return mode
}

func (i *Interpreter) executeShellCmd(shellCmdNode *ts.Node, captureMode shellCaptureMode) shellResult {
	isQuiet := rl.GetChild(shellCmdNode, rl.F_QUIET_MOD) != nil
	isConfirm := rl.GetChild(shellCmdNode, rl.F_CONFIRM_MOD) != nil

	// evaluate the command string
	cmdNode := rl.GetChild(shellCmdNode, rl.F_COMMAND)
	cmdStr := i.eval(cmdNode).Val.
		RequireType(i, cmdNode, "Shell commands must be strings", rl.RadStrT).
		RequireStr(i, shellCmdNode)

	// optional confirmation prompt
	if FlagConfirmShellCommands.Value || isConfirm {
		ok, err := InputConfirm(cmdStr.Plain(), "Run above command? [y/n] > ")
		if err != nil {
			i.errorf(shellCmdNode, "Error confirming shell command: %v", err)
		}
		if !ok {
			empty := bytes.Buffer{}
			return resolveResult(captureMode, empty, empty, 1)
		}
	}

	cmd := resolveCmd(i, shellCmdNode, cmdStr.Plain())
	var stdoutBuf, stderrBuf bytes.Buffer

	// set up output destinations
	if captureMode.captureStdout {
		cmd.Stdout = &stdoutBuf
	} else {
		cmd.Stdout = RIo.StdOut
	}

	if captureMode.captureStderr {
		cmd.Stderr = &stderrBuf
	} else {
		cmd.Stderr = RIo.StdErr
	}

	if !isQuiet {
		RP.RadStderrf(fmt.Sprintf("⚡️ %s\n", cmdStr.String()))
	}

	// Run the command
	err := cmd.Run()

	// handle exit codes and errors
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return resolveResult(captureMode, stdoutBuf, stderrBuf, exitErr.ExitCode())
		}
		i.errorf(shellCmdNode, "Failed to run command: %v\nStderr: %s\n", err, stderrBuf.String())
	}

	return resolveResult(captureMode, stdoutBuf, stderrBuf, 0)
}

func resolveResult(captureMode shellCaptureMode, stdout, stderr bytes.Buffer, exitCode int) shellResult {
	result := shellResult{
		exitCode: exitCode,
	}

	if captureMode.captureStdout {
		s := stdout.String()
		result.stdout = &s
	}

	if captureMode.captureStderr {
		s := stderr.String()
		result.stderr = &s
	}

	return result
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
