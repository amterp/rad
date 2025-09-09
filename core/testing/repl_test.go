package testing

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/amterp/rad/core"
	"github.com/stretchr/testify/assert"
)

// Example of how REPL testing will work once implementation is complete
// This demonstrates the testability architecture using existing RadIo patterns

func replBanner() string {
	return fmt.Sprintf("🤙 Rad REPL %s\nType 'exit()' to quit.\n\n", core.Version)
}

func setupReplTest(t *testing.T, inputs []string) (*bytes.Buffer, *bytes.Buffer) {
	t.Helper()

	// Reset test state
	resetTestState()
	core.IsTest = true

	// Write REPL inputs to stdin buffer
	for _, input := range inputs {
		stdInBuffer.WriteString(input + "\n")
	}

	// Set up args to trigger REPL mode
	args := []string{"--repl"}
	runner := setupRunner(t, args...)

	// Run with REPL flag (will use our injected I/O)
	err := runner.Run()
	assert.NoError(t, err, "REPL should execute without error")

	return stdOutBuffer, stdErrBuffer
}

func TestRepl_BasicSession(t *testing.T) {
	inputs := []string{
		`name = "alice"`,
		`print("Hello {name}!")`,
		`age = 25`,
		`print(age)`,
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	expected := fmt.Sprintf(`🤙 Rad REPL %s
Type 'exit()' to quit.

> > Hello alice!
> > 25
> `, core.Version)

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}

func TestRepl_EnvironmentPersistence(t *testing.T) {
	inputs := []string{
		`x = 10`,
		`y = 20`,
		`print(x + y)`,
		`x = x * 2`,
		`print(x)`,
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	expected := replBanner() + `> > > 30
> > 20
> `

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}

func TestRepl_IncompleteStatementHandling(t *testing.T) {
	inputs := []string{
		`if x > 5:`,
		`x = 10`,
		`print(x)`,
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	// Incomplete statements cause parse errors but REPL continues
	expected := replBanner() + `> Error: Execution error: parse error in statement: if x > 5:
> > 10
> `

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}

func TestRepl_ErrorRecovery(t *testing.T) {
	inputs := []string{
		`x = 5`,
		`y = undefined_variable`, // This should error
		`print(x)`,               // But this should still work
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	// Check the basic REPL functionality and that x variable persists
	output := stdOut.String()
	assert.Contains(t, output, "🤙 Rad REPL", "Should show welcome banner")
	assert.Contains(t, output, "5", "Should print x value after error")

	// Runtime errors go to stderr in Rad, so check that stderr contains error
	errorOutput := stdErr.String()
	assert.Contains(t, errorOutput, "undefined_variable", "Should show undefined variable error in stderr")
	assert.Contains(t, errorOutput, "Undefined variable", "Should show error message in stderr")
}

func TestRepl_EmptyInputHandling(t *testing.T) {
	inputs := []string{
		``,   // Empty line
		`  `, // Whitespace only
		`x = 1`,
		`print(x)`,
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	expected := replBanner() + `> > > > 1
> `

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}

// Test the key feature: expression auto-printing
func TestRepl_ExpressionAutoPrinting(t *testing.T) {
	inputs := []string{
		`2 + 3`,   // Should auto-print
		`10 * 4`,  // Should auto-print
		`x = 100`, // Assignment - no auto-print
		`x`,       // Variable access - should auto-print
		`x / 2`,   // Expression with variable - should auto-print
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	expected := replBanner() + `> 5
> 40
> > 100
> 50
> `

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}

// Test that print statements don't double-print
func TestRepl_PrintStatementHandling(t *testing.T) {
	inputs := []string{
		`x = 42`,
		`print(x)`,      // Should print 42 but not auto-print the return value
		`print("test")`, // Should print test but not auto-print
		`x + 1`,         // Should auto-print 43
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	expected := replBanner() + `> > 42
> test
> 43
> `

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}

// Test complex expressions and function calls
func TestRepl_ComplexExpressions(t *testing.T) {
	inputs := []string{
		`numbers = [1, 2, 3, 4, 5]`,
		`len(numbers)`,
		`sum(numbers)`,
		`numbers[2]`,         // Indexing
		`"hello" + " world"`, // String concatenation
		`exit()`,
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	expected := replBanner() + `> > 5
> 15
> 3
> "hello world"
> `

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}

// Test REPL exit behavior
func TestRepl_ExitBehavior(t *testing.T) {
	inputs := []string{
		`x = 42`,
		`exit()`, // Clean exit
	}

	stdOut, stdErr := setupReplTest(t, inputs)

	expected := replBanner() + `> > `

	assertOutput(t, stdOut, expected)
	assert.Empty(t, stdErr.String())
}
