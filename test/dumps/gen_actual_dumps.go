package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/amterp/rts"
)

// testCase holds one test block from a .dump file.
type testCase struct {
	Title    string
	Code     string
	Expected string
}

// parseTestFile reads a single .dump file and extracts one or more testCase
// blocks. Each block is delimited by lines containing "=====".
//
// State machine notes:
//
//	state = 0 or 2 means we are looking for the "=====" delimiter.
//	state = 1 means we read the next line as the Title.
//	state = 3 means we are gathering lines for Code until the next "=====".
//	state = 4 means we are gathering lines for Expected until the next "=====".
func parseTestFile(path string) ([]testCase, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	const (
		stateDelimiter0 = 0
		stateTitle      = 1
		stateDelimiter1 = 2
		stateCode       = 3
		stateExpected   = 4
	)

	var (
		tests           []testCase
		scanner         = bufio.NewScanner(file)
		state           = stateDelimiter0
		title           string
		codeBuilder     strings.Builder
		expectedBuilder strings.Builder
	)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		switch state {
		case stateDelimiter0, stateDelimiter1:
			// Both 0 and 2 look for a line of "=====", then move on.
			if trimmed == "=====" {
				state++
			}

		case stateTitle:
			// We read exactly one line as the title; then move on to reading Code.
			title = line
			state++

		case stateCode:
			// Gather all code lines until we see "=====" again.
			if trimmed == "=====" {
				state++
				continue
			}
			codeBuilder.WriteString(line)
			codeBuilder.WriteString("\n")

		case stateExpected:
			// Gather all expected lines until we see "=====" again.
			if trimmed == "=====" {
				tests = append(tests, testCase{
					Title:    title,
					Code:     codeBuilder.String(),
					Expected: expectedBuilder.String(),
				})
				// Reset for the next block; back to reading a new Title.
				title = ""
				codeBuilder.Reset()
				expectedBuilder.Reset()
				state = stateTitle
				continue
			}
			expectedBuilder.WriteString(line)
			expectedBuilder.WriteString("\n")
		}
	}

	// If EOF is reached but we have a partially read test, append it.
	if title != "" && (codeBuilder.Len() > 0 || expectedBuilder.Len() > 0) {
		tests = append(tests, testCase{
			Title:    title,
			Code:     codeBuilder.String(),
			Expected: expectedBuilder.String(),
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return tests, nil
}

func parseRsl(code string) (string, error) {
	rslTs, err := rts.NewRslParser()
	if err != nil {
		return "", err
	}
	defer rslTs.Close()

	tree := rslTs.Parse(code)
	return tree.Dump(), nil
}

func main() {
	// 1. Find all .dump files in the cases directory.
	dumpFiles, err := filepath.Glob("./cases/*.dump")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to glob *.dump files: %v", err)
	}

	if len(dumpFiles) == 0 {
		fmt.Fprintf(os.Stderr, "No .dump files found in ./cases/.")
		return
	}

	if err := os.MkdirAll("actual", 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Could not create 'actual' directory: %v", err)
	}

	for _, dumpFile := range dumpFiles {
		testBlocks, err := parseTestFile(dumpFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing %s: %v\n", dumpFile, err)
			continue
		}

		outPath := filepath.Join("actual", filepath.Base(dumpFile))
		file, err := os.Create(outPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", outPath, err)
			continue
		}
		defer file.Close()

		for i, tb := range testBlocks {
			actual, err := parseRsl(tb.Code)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"%s (#%d, %q): Parse error: %v\n",
					dumpFile,
					i+1,
					tb.Title,
					err,
				)
				continue
			}

			content := fmt.Sprintf(
				"=====\n%s\n=====\n%s=====\n%s",
				tb.Title,
				tb.Code,
				actual,
			)

			_, err = file.WriteString(content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error writing actual result to %s: %v\n", outPath, err)
			}
		}
	}
	fmt.Printf("Ran for %d tests dumps.\n", len(dumpFiles))
}
