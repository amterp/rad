package core

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// todo
//  - implement mocking shell responses, like with json requests
//  - colors currently get lost (sometimes?)
//  - tests!
//  - improve error output, especially when stderr is not captured, because that prints first then, before Rad handles it
//  - silent keyword to suppress output?

func (i *MainInterpreter) VisitShellCmdStmt(shellCmd ShellCmd) {
	paths := shellCmd.Paths
	token := shellCmd.Dollar
	if len(paths) > 3 {
		i.error(token, "At most 3 assignments allowed with shell commands")
	}

	cmdValue := shellCmd.CmdExpr.Accept(i)
	cmdStr, ok := cmdValue.(RslString)

	if !ok {
		i.error(token, "Shell command must be a string")
	}

	cmd := resolveCmd(i, token, cmdStr.Plain())
	var stdout, stderr bytes.Buffer

	captureStdout := len(paths) >= 2
	captureStderr := len(paths) >= 3

	var stdoutPipe, stderrPipe io.ReadCloser
	var err error

	if captureStdout {
		stdoutPipe, err = cmd.StdoutPipe()
		if err != nil {
			handleError(i, token, paths, stdout, stderr, 1, fmt.Sprintf("Error creating stdout pipe: %v", err), shellCmd)
		}
	} else {
		cmd.Stdout = RIo.StdOut
	}

	if captureStderr {
		stderrPipe, err = cmd.StderrPipe()
		if err != nil {
			handleError(i, token, paths, stdout, stderr, 1, fmt.Sprintf("Error creating stderr pipe: %v", err), shellCmd)
		}
	} else {
		cmd.Stderr = RIo.StdErr
	}

	if shellCmd.Quiet == nil {
		RP.RadInfo(fmt.Sprintf("⚡️ Running: %s\n", cmdStr.String()))
	}
	if err := cmd.Start(); err != nil {
		handleError(i, token, paths, stdout, stderr, 1, fmt.Sprintf("Error starting command: %v", err), shellCmd)
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
			RP.RadDebug("pipe error")
			handleError(i, token, paths, stdout, stderr, 1, fmt.Sprintf("Failed to run command:\n%s", pipeErr.Error()), shellCmd)
			return
		}
	} else {
		err = cmd.Wait()
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			RP.RadDebug("exit error with error code")
			handleError(i, token, paths, stdout, stderr, exitErr.ExitCode(), fmt.Sprintf("Failed to run command: %v\nStderr: %s", err, stderr.String()), shellCmd)
		} else {
			RP.RadDebug("exit error without error code")
			handleError(i, token, paths, stdout, stderr, 1, fmt.Sprintf("Failed to run command: %v\nStderr: %s", err, stderr.String()), shellCmd)
		}
		return
	}

	doAssignment(i, token, paths, stdout, stderr, 0)
}

func handleError(
	i *MainInterpreter,
	tkn Token,
	paths []VarPath,
	stdout bytes.Buffer,
	stderr bytes.Buffer,
	errorCode int,
	err string,
	cmd ShellCmd,
) {
	doAssignment(i, tkn, paths, stdout, stderr, errorCode)

	if cmd.Bang != nil {
		// Critical command, exit
		RP.ErrorExitCode(err, errorCode)
	}

	if cmd.FailBlock != nil {
		cmd.FailBlock.Accept(i)
		RP.ErrorExitCode(err, errorCode)
	}

	if cmd.RecoverBlock != nil {
		cmd.RecoverBlock.Accept(i)
		return
	}

	if cmd.Unsafe != nil {
		RP.RadInfo(err)
		return
	} else {
		i.error(cmd.Dollar, fmt.Sprintf("Bug! non-unsafe, non-critical shell command without fail or recover block. %s", err))
	}
}

func resolveCmd(i *MainInterpreter, token Token, cmdStr string) *exec.Cmd {
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

	i.error(token, "Cannot run shell cmd as no shell found. Please set the SHELL environment variable.")
	panic(UNREACHABLE)
}

func buildCmd(shellStr string, cmdStr string) *exec.Cmd {
	cmd := exec.Command(shellStr, "-c", cmdStr)
	cmd.Stdin = RIo.StdIn
	//cmd.Stderr = RIo.StdErr // todo ? this seems to conflict with the pipes later
	//cmd.Stdout = RIo.StdOut
	return cmd
}

func doAssignment(
	i *MainInterpreter,
	tkn Token,
	paths []VarPath,
	stdout bytes.Buffer,
	stderr bytes.Buffer,
	errorCode int,
) {
	for j, path := range paths {
		switch j {
		case 0:
			i.setValForPath(tkn, path, int64(errorCode))
		case 1:
			i.setValForPath(tkn, path, stdout.String())
		case 2:
			i.setValForPath(tkn, path, stderr.String())
		}
	}
}
