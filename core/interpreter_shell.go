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
//  - print stdout/stderr as it comes in, unless capturing with identifiers
//    - also re-arrange identifiers so that error code is first. then, if that's all, still don't capture stdout/stderr
//  - implement mocking shell responses, like with json requests
//  - colors currently get lost
//  - tests!
//  - improve error output
//  - quiet keyword

func (i *MainInterpreter) VisitShellCmdStmt(shellCmd ShellCmd) {
	identifiers := shellCmd.Identifiers
	token := shellCmd.Dollar
	if len(identifiers) > 3 {
		i.error(token, "At most 3 identifiers allowed for assignment with shell commands")
	}

	cmdValue := shellCmd.CmdExpr.Accept(i)
	cmdStr, ok := cmdValue.(string)

	if !ok {
		i.error(token, "Shell command must be a string")
	}

	cmd := resolveCmd(i, token, cmdStr)
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		handleError(i, identifiers, stdout, stderr, 1, fmt.Sprintf("Error creating stdout pipe: %v", err), shellCmd.Unsafe, shellCmd.FailureBlock)
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		handleError(i, identifiers, stdout, stderr, 1, fmt.Sprintf("Error creating stderr pipe: %v", err), shellCmd.Unsafe, shellCmd.FailureBlock)
	}

	if err := cmd.Start(); err != nil {
		handleError(i, identifiers, stdout, stderr, 1, fmt.Sprintf("Error starting command: %v", err), shellCmd.Unsafe, shellCmd.FailureBlock)
	}

	errCh := make(chan error, 2) // todo better understand why 2

	// Start goroutine to handle both pipes
	go func() {
		if _, err := io.Copy(&stdout, stdoutPipe); err != nil {
			errCh <- fmt.Errorf("stdout pipe error: %w", err)
			return
		}
		if _, err := io.Copy(&stderr, stderrPipe); err != nil {
			errCh <- fmt.Errorf("stderr pipe error: %w", err)
			return
		}
		errCh <- nil // Signal successful completion
	}()

	err = cmd.Wait()
	if pipeErr := <-errCh; pipeErr != nil {
		handleError(i, identifiers, stdout, stderr, 1, fmt.Sprintf("Hit error:\n%s", pipeErr.Error()), shellCmd.Unsafe, shellCmd.FailureBlock)
		return
	}

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			handleError(i, identifiers, stdout, stderr, exitErr.ExitCode(), fmt.Sprintf("Command failed: %v\nStderr: %s", err, stderr.String()), shellCmd.Unsafe, shellCmd.FailureBlock)
		} else {
			handleError(i, identifiers, stdout, stderr, 1, fmt.Sprintf("Error running command: %v\nStderr: %s", err, stderr.String()), shellCmd.Unsafe, shellCmd.FailureBlock)
		}
		return
	}

	if len(identifiers) == 0 {
		// print stdout
		// todo should really print stdout and stderr in order, and *while* it runs.
		RP.Print(stdout.String())
	} else {
		defineIdentifiers(i, identifiers, stdout, stderr, 0)
	}
}

func handleError(
	i *MainInterpreter,
	identifiers []Token,
	stdout bytes.Buffer,
	stderr bytes.Buffer,
	errorCode int,
	err string,
	unsafe *Token,
	failureBlock *Block,
) {
	defineIdentifiers(i, identifiers, stdout, stderr, errorCode)

	if unsafe == nil {
		if failureBlock != nil {
			failureBlock.Accept(i)
		} else {
			RP.ErrorExitCode(err, errorCode)
		}
	} else {
		// unsafe is defined, in which case we just print the error and move on
		RP.RadInfo(err)
	}
}

func resolveCmd(i *MainInterpreter, token Token, cmdStr string) *exec.Cmd {
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

func defineIdentifiers(i *MainInterpreter, identifiers []Token, stdout bytes.Buffer, stderr bytes.Buffer, errorCode int) {
	for j, identifier := range identifiers {
		switch j {
		case 0:
			i.env.SetAndImplyType(identifier, stdout.String())
		case 1:
			i.env.SetAndImplyType(identifier, stderr.String())
		case 2:
			i.env.SetAndImplyType(identifier, int64(errorCode))
		}
	}
}
