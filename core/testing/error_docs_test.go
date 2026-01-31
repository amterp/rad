package testing

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// isErrorDocFile returns true if the filename matches the error doc naming pattern.
// Error docs are named with their numeric code (e.g., "10001.md", "20000.md").
// This matches the embed pattern `[0-9]*.md` in error_docs.go.
func isErrorDocFile(filename string) bool {
	if !strings.HasSuffix(filename, ".md") {
		return false
	}
	// Must start with a digit
	if len(filename) == 0 {
		return false
	}
	return unicode.IsDigit(rune(filename[0]))
}

// discoverErrorCodes parses rts/rl/errors.go and extracts all error code values.
// This ensures the test automatically fails when new error codes are added without docs.
func discoverErrorCodes(t *testing.T) []string {
	t.Helper()

	errorsFile := "../../rts/rl/errors.go"

	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, errorsFile, nil, parser.ParseComments)
	if err != nil {
		t.Fatalf("Failed to parse %s: %v", errorsFile, err)
	}

	var codes []string

	// Walk the AST looking for const declarations
	ast.Inspect(node, func(n ast.Node) bool {
		// Look for const declarations
		genDecl, ok := n.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			return true
		}

		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			// Check if this is an Error type constant (starts with "Err")
			for i, name := range valueSpec.Names {
				if !strings.HasPrefix(name.Name, "Err") {
					continue
				}

				// Get the value - it should be a string literal like "10001"
				if i < len(valueSpec.Values) {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok && lit.Kind == token.STRING {
						// Remove quotes from the string literal
						code, err := strconv.Unquote(lit.Value)
						if err == nil {
							codes = append(codes, code)
						}
					}
				}
			}
		}
		return true
	})

	if len(codes) == 0 {
		t.Fatal("No error codes found in errors.go - parsing may have failed")
	}

	return codes
}

func TestErrorDocsCompleteness(t *testing.T) {
	errorDocsDir := "../error_docs"

	// Verify the directory exists
	if _, err := os.Stat(errorDocsDir); os.IsNotExist(err) {
		t.Fatalf("Error docs directory not found: %s", errorDocsDir)
	}

	// Discover all error codes from source
	allCodes := discoverErrorCodes(t)

	// Find all existing doc files (only files matching error doc naming pattern)
	existingDocs := make(map[string]bool)
	err := filepath.Walk(errorDocsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if !info.IsDir() && isErrorDocFile(base) {
			code := strings.TrimSuffix(base, ".md")
			existingDocs[code] = true
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk error_docs directory: %v", err)
	}

	// Check which error codes are missing documentation
	var missingDocs []string
	for _, code := range allCodes {
		if !existingDocs[code] {
			missingDocs = append(missingDocs, "RAD"+code)
		}
	}

	if len(missingDocs) > 0 {
		sort.Strings(missingDocs)
		t.Errorf("Missing documentation for %d error codes:\n  %s",
			len(missingDocs), strings.Join(missingDocs, "\n  "))
	}
}

func TestErrorDocsNoOrphans(t *testing.T) {
	errorDocsDir := "../error_docs"

	// Discover all error codes from source
	allCodes := discoverErrorCodes(t)

	// Build a map of all known error codes
	knownCodes := make(map[string]bool)
	for _, code := range allCodes {
		knownCodes[code] = true
	}

	// Find doc files that don't correspond to any known error code
	// Only check files matching error doc naming pattern (starting with digits)
	var orphanDocs []string
	err := filepath.Walk(errorDocsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if !info.IsDir() && isErrorDocFile(base) {
			code := strings.TrimSuffix(base, ".md")
			if !knownCodes[code] {
				orphanDocs = append(orphanDocs, "RAD"+code)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("Failed to walk error_docs directory: %v", err)
	}

	if len(orphanDocs) > 0 {
		sort.Strings(orphanDocs)
		t.Errorf("Found documentation for unknown error codes:\n  %s",
			strings.Join(orphanDocs, "\n  "))
	}
}
