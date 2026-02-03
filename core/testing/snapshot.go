package testing

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

// UpdateSnapshots is set by the -update flag to regenerate expected outputs
var UpdateSnapshots = flag.Bool("update", false, "update snapshot expected outputs")

const (
	TitleDelimiter       = "### TITLE ###"
	DescriptionDelimiter = "### DESCRIPTION ###"
	SkipDelimiter        = "### SKIP ###"
	InputDelimiter       = "### INPUT ###"
	ArgsDelimiter        = "### ARGS ###"
	RawArgsDelimiter     = "### RAW_ARGS ###"
	StdoutDelimiter      = "### STDOUT ###"
	StderrDelimiter      = "### STDERR ###"
	ExitDelimiter        = "### EXIT ###"
)

// SnapshotCase holds one test case from a snapshot file.
type SnapshotCase struct {
	Title       string
	Description string // Optional documentation/comments about the test
	Input       string
	Args        []string // Command-line arguments to pass to the script

	// Skip support - if SkipReason is non-empty, the test is skipped
	SkipReason string

	// RawArgs suppresses automatic flag additions (like --color=never).
	// Use ### RAW_ARGS ### instead of ### ARGS ### to enable this.
	RawArgs bool

	Stdout   string
	Stderr   string
	ExitCode int
}

// ParseSnapshotFile reads a .snap file and extracts test cases.
//
// Format:
//
//	### TITLE ###
//	<test name>
//	### DESCRIPTION ###
//	<optional documentation, can be multi-line>
//	### SKIP ###
//	<optional skip reason - if section present, test is skipped>
//	### INPUT ###
//	<code to run>
//	### ARGS ###         (or ### RAW_ARGS ### to suppress auto --color=never)
//	<args, one per line> (optional)
//	### STDOUT ###
//	<expected stdout> (optional, omit section for empty)
//	### STDERR ###
//	<expected stderr> (optional, omit section for empty)
//	### EXIT ###
//	<exit code> (optional, defaults to 0)
//
// Multiple test cases can be included by repeating the pattern.
func ParseSnapshotFile(path string) ([]SnapshotCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	const (
		stateInit        = iota // Looking for ### TITLE ###
		stateTitle              // Reading title line
		stateDescription        // Reading description lines
		stateSkip               // Reading skip reason
		stateInput              // Reading input lines
		stateArgs               // Reading args lines
		stateStdout             // Reading stdout lines
		stateStderr             // Reading stderr lines
		stateExit               // Reading exit code
	)

	var (
		cases              []SnapshotCase
		scanner            = bufio.NewScanner(file)
		state              = stateInit
		lineNum            = 0
		title              string
		descriptionBuilder strings.Builder
		skipReason         strings.Builder
		inputBuilder       strings.Builder
		args               []string
		rawArgs            bool
		stdoutBuilder      strings.Builder
		stderrBuilder      strings.Builder
		exitCode           int
		exitCodeSet        bool
	)

	finishCase := func() {
		if title != "" || inputBuilder.Len() > 0 {
			tc := SnapshotCase{
				Title:       title,
				Description: strings.TrimSuffix(descriptionBuilder.String(), "\n"),
				SkipReason:  strings.TrimSuffix(skipReason.String(), "\n"),
				Input:       strings.TrimSuffix(inputBuilder.String(), "\n"),
				Args:        args,
				RawArgs:     rawArgs,
				// Trim one trailing newline - it's the format separator, not content.
				// If output truly ends with newline(s), they appear as blank lines
				// before the next section header.
				Stdout:   strings.TrimSuffix(stdoutBuilder.String(), "\n"),
				Stderr:   strings.TrimSuffix(stderrBuilder.String(), "\n"),
				ExitCode: exitCode,
			}

			cases = append(cases, tc)
		}

		// Reset for next case
		title = ""
		descriptionBuilder.Reset()
		skipReason.Reset()
		inputBuilder.Reset()
		args = nil
		rawArgs = false
		stdoutBuilder.Reset()
		stderrBuilder.Reset()
		exitCode = 0
		exitCodeSet = false
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
			// After title, expect DESCRIPTION, SKIP, or INPUT
			if scanner.Scan() {
				lineNum++
				nextLine := strings.TrimSpace(scanner.Text())
				switch nextLine {
				case DescriptionDelimiter:
					state = stateDescription
				case SkipDelimiter:
					state = stateSkip
				case InputDelimiter:
					state = stateInput
				default:
					return nil, fmt.Errorf("%s:%d: expected '%s', '%s', or '%s' after title, got '%s'",
						path, lineNum, DescriptionDelimiter, SkipDelimiter, InputDelimiter, nextLine)
				}
			}

		case stateDescription:
			switch trimmed {
			case SkipDelimiter:
				state = stateSkip
			case InputDelimiter:
				state = stateInput
			case TitleDelimiter:
				finishCase()
				state = stateTitle
			default:
				descriptionBuilder.WriteString(line)
				descriptionBuilder.WriteString("\n")
			}

		case stateSkip:
			switch trimmed {
			case InputDelimiter:
				state = stateInput
			case TitleDelimiter:
				finishCase()
				state = stateTitle
			default:
				skipReason.WriteString(line)
				skipReason.WriteString("\n")
			}

		case stateInput:
			switch trimmed {
			case ArgsDelimiter:
				state = stateArgs
			case RawArgsDelimiter:
				rawArgs = true
				state = stateArgs
			case StdoutDelimiter:
				state = stateStdout
			case StderrDelimiter:
				state = stateStderr
			case ExitDelimiter:
				state = stateExit
			case TitleDelimiter:
				// Empty test case, start new one
				finishCase()
				state = stateTitle
			default:
				inputBuilder.WriteString(line)
				inputBuilder.WriteString("\n")
			}

		case stateArgs:
			switch trimmed {
			case StdoutDelimiter:
				state = stateStdout
			case StderrDelimiter:
				state = stateStderr
			case ExitDelimiter:
				state = stateExit
			case TitleDelimiter:
				finishCase()
				state = stateTitle
			default:
				// Each non-empty line is an argument
				if trimmed != "" {
					args = append(args, line)
				}
			}

		case stateStdout:
			switch trimmed {
			case StderrDelimiter:
				state = stateStderr
			case ExitDelimiter:
				state = stateExit
			case TitleDelimiter:
				finishCase()
				state = stateTitle
			default:
				stdoutBuilder.WriteString(line)
				stdoutBuilder.WriteString("\n")
			}

		case stateStderr:
			switch trimmed {
			case StdoutDelimiter:
				state = stateStdout
			case ExitDelimiter:
				state = stateExit
			case TitleDelimiter:
				finishCase()
				state = stateTitle
			default:
				stderrBuilder.WriteString(line)
				stderrBuilder.WriteString("\n")
			}

		case stateExit:
			// Exit code should be on one line
			if trimmed == TitleDelimiter {
				finishCase()
				state = stateTitle
			} else if trimmed != "" && !exitCodeSet {
				code, err := strconv.Atoi(trimmed)
				if err != nil {
					return nil, fmt.Errorf("%s:%d: invalid exit code '%s': %w",
						path, lineNum, trimmed, err)
				}
				exitCode = code
				exitCodeSet = true
			}

		}
	}

	// Finish the last case
	if state != stateInit && state != stateTitle {
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
		// Add separator between cases, but only if the previous content doesn't end with a newline
		if i > 0 {
			str := builder.String()
			if len(str) > 0 && str[len(str)-1] != '\n' {
				builder.WriteString("\n")
			}
		}
		builder.WriteString(TitleDelimiter)
		builder.WriteString("\n")
		builder.WriteString(tc.Title)
		builder.WriteString("\n")

		if tc.Description != "" {
			builder.WriteString(DescriptionDelimiter)
			builder.WriteString("\n")
			builder.WriteString(tc.Description)
			builder.WriteString("\n")
		}

		if tc.SkipReason != "" {
			builder.WriteString(SkipDelimiter)
			builder.WriteString("\n")
			builder.WriteString(tc.SkipReason)
			builder.WriteString("\n")
		}

		builder.WriteString(InputDelimiter)
		builder.WriteString("\n")
		builder.WriteString(tc.Input)
		builder.WriteString("\n")

		// Only write sections that have content or non-default values
		if len(tc.Args) > 0 || tc.RawArgs {
			if tc.RawArgs {
				builder.WriteString(RawArgsDelimiter)
			} else {
				builder.WriteString(ArgsDelimiter)
			}
			builder.WriteString("\n")
			for _, arg := range tc.Args {
				builder.WriteString(arg)
				builder.WriteString("\n")
			}
		}
		if tc.Stdout != "" {
			builder.WriteString(StdoutDelimiter)
			builder.WriteString("\n")
			builder.WriteString(tc.Stdout)
			builder.WriteString("\n")
		}
		if tc.Stderr != "" {
			builder.WriteString(StderrDelimiter)
			builder.WriteString("\n")
			builder.WriteString(tc.Stderr)
			builder.WriteString("\n")
		}
		if tc.ExitCode != 0 {
			builder.WriteString(ExitDelimiter)
			builder.WriteString("\n")
			builder.WriteString(strconv.Itoa(tc.ExitCode))
			builder.WriteString("\n")
		}
	}

	return os.WriteFile(path, []byte(builder.String()), 0644)
}

// SnapshotResult holds the actual output from running a test case.
type SnapshotResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

// CompareSnapshotResult compares actual output against expected.
// Returns true if the snapshot needs updating.
// In update mode, it does not fail the test on mismatch.
func CompareSnapshotResult(t *testing.T, tc *SnapshotCase, actual SnapshotResult) bool {
	t.Helper()

	needsUpdate := false

	// Compare stdout
	if actual.Stdout != tc.Stdout {
		needsUpdate = true
		if !*UpdateSnapshots {
			t.Errorf("Stdout mismatch\nExpected:\n%q\nActual:\n%q", tc.Stdout, actual.Stdout)
		}
	}

	// Compare stderr
	if actual.Stderr != tc.Stderr {
		needsUpdate = true
		if !*UpdateSnapshots {
			t.Errorf("Stderr mismatch\nExpected:\n%q\nActual:\n%q", tc.Stderr, actual.Stderr)
		}
	}

	// Compare exit code
	if actual.ExitCode != tc.ExitCode {
		needsUpdate = true
		if !*UpdateSnapshots {
			t.Errorf("Exit code mismatch\nExpected: %d\nActual: %d", tc.ExitCode, actual.ExitCode)
		}
	}

	return needsUpdate
}
