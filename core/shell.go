package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

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
	leftVarPathNodes := i.getChildren(shellStmtNode, F_LEFT)
	numExpectedOutputs := len(leftVarPathNodes)

	if numExpectedOutputs > 3 {
		i.errorf(shellStmtNode, "At most 3 assignments allowed with shell commands")
	}

	shellCmdNode := i.getChild(shellStmtNode, F_SHELL_CMD)
	result := i.executeShellCmd(shellCmdNode, len(leftVarPathNodes))

	if numExpectedOutputs >= 1 {
		i.doVarPathAssign(&leftVarPathNodes[0], newRslValue(i, shellCmdNode, result.exitCode))
	}
	if numExpectedOutputs >= 2 {
		i.doVarPathAssign(&leftVarPathNodes[1], newRslValue(i, shellCmdNode, *result.stdout))
	}
	if numExpectedOutputs >= 3 {
		i.doVarPathAssign(&leftVarPathNodes[2], newRslValue(i, shellCmdNode, *result.stderr))
	}

	if result.exitCode != 0 {
		stmtNodes := i.getChildren(shellCmdNode, F_STMT)
		i.runBlock(stmtNodes)
		responseNode := i.getChild(shellCmdNode, F_RESPONSE)
		if responseNode != nil {
			if responseNode.Kind() == K_FAIL {
				RP.ErrorExitCode("", result.exitCode)
			}
		}
	}
}

func (i *Interpreter) executeShellCmd(shellCmdNode *ts.Node, numExpectedOutputs int) shellResult {
	isQuiet := i.getChild(shellCmdNode, F_QUIET_MOD) != nil

	cmdNode := i.getChild(shellCmdNode, F_COMMAND)
	cmdStr := i.evaluate(cmdNode, 1)[0].
		RequireType(i, cmdNode, "Shell commands must be strings", RslStringT).
		RequireStr(i, shellCmdNode)

	if FlagConfirmShellCommands.Value {
		prompt := fmt.Sprintf("%s\nRun above command? [y/n] > ", cmdStr.Plain())
		response, err := InteractiveConfirm(prompt)
		if err != nil {
			i.errorf(shellCmdNode, "Error confirming shell command: %v", err)
		}
		if !response {
			i.errorf(shellCmdNode, RED.Colorize("Aborted shell command, exiting."))
		}
	}

	cmd := resolveCmd(i, shellCmdNode, cmdStr.Plain())
	var stdout, stderr bytes.Buffer

	captureStdout := numExpectedOutputs >= 2
	captureStderr := numExpectedOutputs >= 3

	var stdoutPipe, stderrPipe io.ReadCloser
	var err error

	if captureStdout {
		stdoutPipe, err = cmd.StdoutPipe()
		if err != nil {
			i.errorf(shellCmdNode, "Error creating stdout pipe: %v", err)
		}
	} else {
		cmd.Stdout = RIo.StdOut
	}

	if captureStderr {
		stderrPipe, err = cmd.StderrPipe()
		if err != nil {
			i.errorf(shellCmdNode, "Error creating stderr pipe: %v", err)
		}
	} else {
		cmd.Stderr = RIo.StdErr
	}

	if !isQuiet {
		RP.RadInfo(fmt.Sprintf("⚡️ Running: %s\n", cmdStr.String()))
	}
	if err = cmd.Start(); err != nil {
		i.errorf(shellCmdNode, "Error starting command: %v", err)
	}

	if captureStdout || captureStderr {
		errCh := make(chan error, 2)

		go func() {
			if captureStdout {
				if _, err := io.Copy(&stdout, stdoutPipe); err != nil {
					errCh <- fmt.Errorf("stdout pipe error: %w", err)
					return
				}
			}
			if captureStderr {
				if _, err := io.Copy(&stderr, stderrPipe); err != nil {
					errCh <- fmt.Errorf("stderr pipe error: %w", err)
					return
				}
			}
			errCh <- nil
		}()

		err = cmd.Wait()
		if pipeErr := <-errCh; pipeErr != nil {
			RP.RadDebugf("pipe error")
			i.errorf(shellCmdNode, "Failed to run command: %s", pipeErr)
		}
	} else {
		err = cmd.Wait()
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			// cmd return non-0, which is valid in rad
			RP.RadDebugf("exit error with error code")
			return resolveResult(shellCmdNode, numExpectedOutputs, stdout, stderr, exitErr.ExitCode())
		} else {
			// genuine error, error exit no matter what
			RP.RadDebugf("exit error without error code")
			i.errorf(shellCmdNode, "Failed to run command: %v\nStderr: %s\n", err, stderr.String())
		}
	}

	return resolveResult(shellCmdNode, numExpectedOutputs, stdout, stderr, 0)
}

func resolveResult(shellNode *ts.Node, numExpectedOutputs int, stdout, stderr bytes.Buffer, exitCode int) shellResult {
	isCritical := shellNode.Kind() == K_CRITICAL_SHELL_CMD
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
	cmd.Stdin = RIo.StdIn
	//cmd.Stderr = RIo.StdErr // todo ? this seems to conflict with the pipes later
	//cmd.Stdout = RIo.StdOut
	return cmd
}
