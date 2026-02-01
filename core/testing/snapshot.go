package testing

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"
	"testing"
)

// UpdateSnapshots is set by the -update flag to regenerate expected outputs
var UpdateSnapshots = flag.Bool("update", false, "update snapshot expected outputs")

const (
	TitleDelimiter    = "### TITLE ###"
	InputDelimiter    = "### INPUT ###"
	ExpectedDelimiter = "### EXPECTED ###"
)

// SnapshotCase holds one test case from a snapshot file.
type SnapshotCase struct {
	Title    string
	Input    string
	Expected string
}

// ParseSnapshotFile reads a .snap file and extracts test cases.
// The file format is:
//
//	### TITLE ###
//	<test name>
//	### INPUT ###
//	<code to parse/run>
//	### EXPECTED ###
//	<expected output>
//
// Multiple test cases can be included by repeating the pattern.
func ParseSnapshotFile(path string) ([]SnapshotCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	const (
		stateInit     = iota // Looking for ### TITLE ###
		stateTitle           // Reading title line
		stateInput           // Reading input lines until ### EXPECTED ###
		stateExpected        // Reading expected lines until ### TITLE ### or EOF
	)

	var (
		cases           []SnapshotCase
		scanner         = bufio.NewScanner(file)
		state           = stateInit
		lineNum         = 0
		title           string
		inputBuilder    strings.Builder
		expectedBuilder strings.Builder
	)

	finishCase := func() {
		if title != "" || inputBuilder.Len() > 0 || expectedBuilder.Len() > 0 {
			cases = append(cases, SnapshotCase{
				Title:    title,
				Input:    strings.TrimSuffix(inputBuilder.String(), "\n"),
				Expected: expectedBuilder.String(),
			})
			title = ""
			inputBuilder.Reset()
			expectedBuilder.Reset()
		}
	}

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch state {
		case stateInit:
			if trimmed == TitleDelimiter {
				state = stateTitle
			}

		case stateTitle:
			title = line
			state = stateInput
			// Expect ### INPUT ### on the next line
			if scanner.Scan() {
				lineNum++
				nextLine := strings.TrimSpace(scanner.Text())
				if nextLine != InputDelimiter {
					return nil, fmt.Errorf("%s:%d: expected '%s' after title, got '%s'",
						path, lineNum, InputDelimiter, nextLine)
				}
			}

		case stateInput:
			if trimmed == ExpectedDelimiter {
				state = stateExpected
				continue
			}
			inputBuilder.WriteString(line)
			inputBuilder.WriteString("\n")

		case stateExpected:
			if trimmed == TitleDelimiter {
				finishCase()
				state = stateTitle
				continue
			}
			expectedBuilder.WriteString(line)
			expectedBuilder.WriteString("\n")
		}
	}

	// Finish the last case if we have one
	if state == stateExpected {
		finishCase()
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return cases, nil
}

// WriteSnapshotFile writes test cases back to a .snap file.
func WriteSnapshotFile(path string, cases []SnapshotCase) error {
	var builder strings.Builder

	for i, tc := range cases {
		if i > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(TitleDelimiter)
		builder.WriteString("\n")
		builder.WriteString(tc.Title)
		builder.WriteString("\n")
		builder.WriteString(InputDelimiter)
		builder.WriteString("\n")
		builder.WriteString(tc.Input)
		builder.WriteString("\n")
		builder.WriteString(ExpectedDelimiter)
		builder.WriteString("\n")
		builder.WriteString(tc.Expected)
	}

	return os.WriteFile(path, []byte(builder.String()), 0644)
}

// CompareSnapshot compares actual output against expected.
// Returns true if the snapshot needs updating (actual differs from expected).
// In update mode, it does not fail the test on mismatch.
// The caller is responsible for updating tc.Expected when this returns true.
func CompareSnapshot(t *testing.T, tc *SnapshotCase, actual string) bool {
	t.Helper()

	if actual != tc.Expected {
		if *UpdateSnapshots {
			return true // Caller should update tc.Expected
		}
		t.Errorf("Snapshot mismatch\nExpected:\n%s\nActual:\n%s", tc.Expected, actual)
	}
	return false
}
