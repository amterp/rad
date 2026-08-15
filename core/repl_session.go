package core

import (
	"io"
	"strings"
)

// Core REPL contracts and interfaces following maintainability-first design

// ReplSession represents the main REPL session contract
type ReplSession interface {
	Run() error
	ExecuteStatement(input string) (*ExecutionResult, error)
	GetEnvironment() *Env
	Shutdown() error
}

// InputReader abstracts input handling (designed for future multi-line extension)
type InputReader interface {
	ReadStatement() (string, error)
	SupportsMultiLine() bool
	SetPrompt(primary, continuation string)
	Shutdown() error
}

// ExecutionResult represents the result of executing a statement
type ExecutionResult struct {
	Value       RadValue
	ShouldPrint bool
	Error       *RadError
}

// NewExecutionResult creates a new execution result
func NewExecutionResult(value RadValue, shouldPrint bool, err *RadError) *ExecutionResult {
	return &ExecutionResult{
		Value:       value,
		ShouldPrint: shouldPrint,
		Error:       err,
	}
}

// DefaultReplSession implements the ReplSession interface
type DefaultReplSession struct {
	interpreter *Interpreter
	inputReader InputReader
	// todo: Add session state (history, etc.)
}

// NewReplSession creates a new REPL session with the given interpreter and input reader
func NewReplSession(interpreter *Interpreter, inputReader InputReader) ReplSession {
	return &DefaultReplSession{
		interpreter: interpreter,
		inputReader: inputReader,
	}
}

// Run starts the main REPL loop
func (s *DefaultReplSession) Run() error {
	printWelcomeBanner()

	// Set default prompts
	s.inputReader.SetPrompt("> ", "... ")

	// Main read-eval-print loop
	for {
		// Read input from user
		input, err := s.inputReader.ReadStatement()
		if err != nil {
			handleReplError(err)
			if err == io.EOF {
				// Ctrl+D - clean exit
				break
			}
			continue
		}

		// Skip empty input
		if strings.TrimSpace(input) == "" {
			continue
		}

		// Check for exit command
		if strings.TrimSpace(input) == "exit()" {
			break
		}

		// Execute the statement
		result, err := s.ExecuteStatement(input)
		if err != nil {
			handleReplError(err)
			continue
		}

		// Handle execution result
		if result.Error != nil {
			// result.Error is a *RadError, format it properly
			RP.Printf("Error: %v\n", result.Error.Msg().Plain())
		} else if shouldPrintResult(result) {
			RP.Printf("%s\n", ToPrintable(result.Value))
		}
	}

	return nil
}

// ExecuteStatement executes a single statement and returns the result
func (s *DefaultReplSession) ExecuteStatement(input string) (result *ExecutionResult, err error) {
	// A fatal error unwinds as *RadAbort instead of ending the process, and the
	// turn is where it stops: the diagnostic is already on screen, so the
	// session just carries on at the next prompt. The exit latch is cleared so
	// the following turn's deferred statements still run.
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if _, ok := r.(*RadAbort); !ok {
			panic(r)
		}
		RExit.ResetExiting()
		result, err = NewExecutionResult(VOID_SENTINEL, false, nil), nil
	}()

	resultValue, evalErr := s.interpreter.EvaluateStatement(input)
	if evalErr != nil {
		radErr := NewErrorStrf("Execution error: %v", evalErr)
		return NewExecutionResult(RAD_NULL_VAL, false, radErr), nil
	}

	// Expressions print their value; assignments and void calls print nothing.
	return NewExecutionResult(resultValue, resultValue != VOID_SENTINEL, nil), nil
}

// GetEnvironment returns the current interpreter environment
func (s *DefaultReplSession) GetEnvironment() *Env {
	return s.interpreter.env
}

// Shutdown performs cleanup when REPL session ends
func (s *DefaultReplSession) Shutdown() error {
	// Clean up input reader
	if err := s.inputReader.Shutdown(); err != nil {
		return err
	}

	// todo: (maybe) save command history to file

	return nil
}
