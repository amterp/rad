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
func (s *DefaultReplSession) ExecuteStatement(input string) (*ExecutionResult, error) {
	// Use the new EvaluateStatement API on our persistent interpreter
	resultValue, err := s.interpreter.EvaluateStatement(input)
	if err != nil {
		// Convert Go error to RadError for consistent error handling
		radErr := NewErrorStrf("Execution error: %v", err)
		return NewExecutionResult(RAD_NULL_VAL, false, radErr), nil
	}

	// Determine if result should be printed
	// For MVP: print expressions that return values, don't print assignments or print statements

	shouldPrint := resultValue != VOID_SENTINEL

	return NewExecutionResult(resultValue, shouldPrint, nil), nil
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
