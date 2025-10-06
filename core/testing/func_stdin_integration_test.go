package testing

import (
	"os"
	"testing"
)

func Test_Stdin_Integration_WC_Basic(t *testing.T) {
	// Read the test input file
	inputData, err := os.ReadFile("data/wc_input.txt")
	if err != nil {
		t.Fatalf("Failed to read test input: %v", err)
	}

	// Create test params with script file and stdin input
	tp := NewTestParams("", "./rad_scripts/wc.rad", "--color=never").StdinInput(string(inputData))
	setupAndRun(t, tp)

	// The wc_input.txt file has:
	// - 3 lines
	// - 11 words
	// - 57 characters (including newlines)
	assertOnlyOutput(t, stdOutBuffer, "3 11 57\n")
	assertNoErrors(t)
}

func Test_Stdin_Integration_WC_LinesOnly(t *testing.T) {
	// Read the test input file
	inputData, err := os.ReadFile("data/wc_input.txt")
	if err != nil {
		t.Fatalf("Failed to read test input: %v", err)
	}

	// Test with -l flag
	tp := NewTestParams("", "./rad_scripts/wc.rad", "--color=never", "-l").StdinInput(string(inputData))
	setupAndRun(t, tp)

	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Stdin_Integration_WC_LinesOnlyLongFlag(t *testing.T) {
	// Read the test input file
	inputData, err := os.ReadFile("data/wc_input.txt")
	if err != nil {
		t.Fatalf("Failed to read test input: %v", err)
	}

	// Test with --lines flag
	tp := NewTestParams("", "./rad_scripts/wc.rad", "--color=never", "--lines").StdinInput(string(inputData))
	setupAndRun(t, tp)

	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_Stdin_Integration_WC_NoStdin(t *testing.T) {
	// Test error when no stdin provided
	tp := NewTestParams("", "./rad_scripts/wc.rad", "--color=never")
	setupAndRun(t, tp)

	assertOutput(t, stdErrBuffer, "wc: no input\n")
	assertExitCode(t, 1)
}

func Test_Stdin_Integration_WC_EmptyStdin(t *testing.T) {
	// Test with empty stdin
	tp := NewTestParams("", "./rad_scripts/wc.rad", "--color=never").StdinInput("")
	setupAndRun(t, tp)

	// Empty input should give 1 empty line, 0 words, 0 chars
	assertOnlyOutput(t, stdOutBuffer, "1 0 0\n")
	assertNoErrors(t)
}
