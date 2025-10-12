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
//  - colors currently get lost (sometimes?)
//  - tests!
//  - improve error output, especially when stderr is not captured, because that prints first then, before Rad handles it
//  - silent keyword to suppress output?

type shellResult struct {
	exitCode int
	stdout   *string
	stderr   *string
}

func (i *Interpreter) executeShellStmt(shellStmtNode *ts.Node) {
	leftNode := rl.GetChildren(shellStmtNode, rl.F_LEFT)
	leftNodes := rl.GetChildren(shellStmtNode, rl.F_LEFTS)
	leftNodes = append(leftNode, leftNodes...) // todo hacky

	numExpectedOutputs := len(leftNodes)

	if numExpectedOutputs > 3 {
		i.errorf(shellStmtNode, "At most 3 assignments allowed with shell commands")
	}

	shellCmdNode := rl.GetChild(shellStmtNode, rl.F_SHELL_CMD)
	result := i.executeShellCmd(shellCmdNode, numExpectedOutputs)

	if numExpectedOutputs >= 1 {
		i.doVarPathAssign(&leftNodes[0], newRadValue(i, shellCmdNode, result.exitCode), false)
	}
	if numExpectedOutputs >= 2 {
		i.doVarPathAssign(&leftNodes[1], newRadValue(i, shellCmdNode, *result.stdout), false)
	}
	if numExpectedOutputs >= 3 {
		i.doVarPathAssign(&leftNodes[2], newRadValue(i, shellCmdNode, *result.stderr), false)
	}

	if result.exitCode != 0 {
		stmtNodes := rl.GetChildren(shellCmdNode, rl.F_STMT)
		i.runBlock(stmtNodes)
		responseNode := rl.GetChild(shellCmdNode, rl.F_RESPONSE)
		if responseNode != nil {
			if responseNode.Kind() == rl.K_FAIL {
				RP.ErrorExitCode("", result.exitCode)
			}
		}
	}
}

func (i *Interpreter) executeShellCmd(shellCmdNode *ts.Node, numExpectedOutputs int) shellResult {
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
			return resolveResult(shellCmdNode, numExpectedOutputs, empty, empty, 1)
		}
	}

	cmd := resolveCmd(i, shellCmdNode, cmdStr.Plain())
	var stdoutBuf, stderrBuf bytes.Buffer

	captureStdout := numExpectedOutputs >= 2
	captureStderr := numExpectedOutputs >= 3

	// set up output destinations
	if captureStdout {
		cmd.Stdout = &stdoutBuf
	} else {
		cmd.Stdout = RIo.StdOut
	}

	if captureStderr {
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
			return resolveResult(shellCmdNode, numExpectedOutputs, stdoutBuf, stderrBuf, exitErr.ExitCode())
		}
		i.errorf(shellCmdNode, "Failed to run command: %v\nStderr: %s\n", err, stderrBuf.String())
	}

	return resolveResult(shellCmdNode, numExpectedOutputs, stdoutBuf, stderrBuf, 0)
}
func resolveResult(shellNode *ts.Node, numExpectedOutputs int, stdout, stderr bytes.Buffer, exitCode int) shellResult {
	isCritical := shellNode.Kind() == rl.K_CRITICAL_SHELL_CMD
	if isCritical && exitCode != 0 {
		RP.ErrorCodeExitf(exitCode, stderr.String())
	}

	result := shellResult{
		exitCode: exitCode,
	}

	if numExpectedOutputs > 1 {
		s := stdout.String()
		result.stdout = &s
	}

	if numExpectedOutputs > 2 {
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
