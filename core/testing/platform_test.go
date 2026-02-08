package testing

import (
	"os"
	"strings"
	"testing"
)

// These tests verify that all path-returning functions use forward slashes
// regardless of platform. This is critical for script portability.

func Test_Platform_GetPathReturnsNormalizedPath(t *testing.T) {
	script := `
p = get_path("data/test_file.txt")
print(p.full_path)
`
	setupAndRunCode(t, script, "--color=never")
	output := stdOutBuffer.String()
	// Path should use forward slashes, not backslashes
	if strings.Contains(output, "\\") {
		t.Errorf("get_path().full_path contains backslashes: %s", output)
	}
	assertNoErrors(t)
}

func Test_Platform_GetRadHomeReturnsNormalizedPath(t *testing.T) {
	script := `
home = get_rad_home()
print(home)
`
	setupAndRunCode(t, script, "--color=never")
	output := stdOutBuffer.String()
	// Path should use forward slashes, not backslashes
	if strings.Contains(output, "\\") {
		t.Errorf("get_rad_home() contains backslashes: %s", output)
	}
	assertNoErrors(t)
}

func Test_Platform_GetStashPathReturnsNormalizedPath(t *testing.T) {
	script := `
---
@stash_id = test_id
---
dir = get_stash_path()
print(dir)
`
	setupAndRunCode(t, script, "--color=never")
	output := stdOutBuffer.String()
	// Path should use forward slashes, not backslashes
	if strings.Contains(output, "\\") {
		t.Errorf("get_stash_path() contains backslashes: %s", output)
	}
	assertNoErrors(t)
}

func Test_Platform_GetStashPathSubPathReturnsNormalizedPath(t *testing.T) {
	script := `
---
@stash_id = test_id
---
dir = get_stash_path("some/sub/path.txt")
print(dir)
`
	setupAndRunCode(t, script, "--color=never")
	output := stdOutBuffer.String()
	// Path should use forward slashes, not backslashes
	if strings.Contains(output, "\\") {
		t.Errorf("get_stash_path(subpath) contains backslashes: %s", output)
	}
	assertNoErrors(t)
}

func Test_Platform_FindPathsReturnsNormalizedPaths(t *testing.T) {
	script := `
paths = find_paths("path_example/dir2")
for p in paths:
	print(p)
`
	setupAndRunCode(t, script, "--color=never")
	output := stdOutBuffer.String()
	// All paths should use forward slashes, not backslashes
	if strings.Contains(output, "\\") {
		t.Errorf("find_paths() returned paths with backslashes: %s", output)
	}
	assertNoErrors(t)
}

func Test_Platform_LoadStashFileReturnsNormalizedPath(t *testing.T) {
	script := `
---
@stash_id = with_stash
---
r = load_stash_file("existing.txt", "default")
print(r.full_path)
`
	setupAndRunCode(t, script, "--color=never")
	output := stdOutBuffer.String()
	// Path should use forward slashes, not backslashes
	if strings.Contains(output, "\\") {
		t.Errorf("load_stash_file().full_path contains backslashes: %s", output)
	}
	assertNoErrors(t)
}

func Test_Platform_WriteFileReturnsNormalizedPath(t *testing.T) {
	script := `
result = write_file("data/test_write_platform.txt", "test")
print(result.path)
delete_path("data/test_write_platform.txt")
`
	setupAndRunCode(t, script, "--color=never")
	output := stdOutBuffer.String()
	// Path should use forward slashes, not backslashes
	if strings.Contains(output, "\\") {
		t.Errorf("write_file().path contains backslashes: %s", output)
	}
	assertNoErrors(t)
}

// Test that read_file normalizes line endings in text mode
func Test_Platform_ReadFileNormalizesLineEndings(t *testing.T) {
	// Create a test file with actual Windows line endings (CRLF) using Go
	testFile := "data/crlf_test.txt"
	crlfContent := []byte("line1\r\nline2\r\nline3")
	if err := os.WriteFile(testFile, crlfContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	// Read it back via Rad - should have normalized line endings
	script := `
result = read_file("data/crlf_test.txt")
// Original: "line1\r\nline2\r\nline3" = 5+2+5+2+5 = 19 chars
// After normalization: "line1\nline2\nline3" = 5+1+5+1+5 = 17 chars
lines = result.content.split("\n")
print(len(lines))
print(len(result.content))
`
	setupAndRunCode(t, script, "--color=never")
	// 3 lines, 17 characters after normalization (19 - 2 CRs removed)
	expected := `3
17
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// Test that paths can be split with "/" on all platforms
func Test_Platform_PathsSplittableWithForwardSlash(t *testing.T) {
	script := `
p = get_path("data/test_file.txt")
parts = p.full_path.split("/")
// Last part should be the filename
print(parts[-1])
`
	setupAndRunCode(t, script, "--color=never")
	expected := `test_file.txt
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
