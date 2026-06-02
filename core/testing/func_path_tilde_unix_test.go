//go:build unix

package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// Issue 125: path builtins should expand a leading "~" consistently, matching
// shell behavior. These tests point $HOME at a temp dir (os.UserHomeDir reads
// $HOME on unix) so "~" resolves there without touching the real home.

func Test_PathTilde_WriteReadDeleteRoundTrip(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	script := `
w = write_file("~/issue125.txt", "hi there")
print(w.path)
print(read_file("~/issue125.txt").content)
print(delete_path("~/issue125.txt"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := fmt.Sprintf("%s/issue125.txt\nhi there\ntrue\n", filepath.ToSlash(tmpHome))
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	if _, err := os.Stat(filepath.Join(tmpHome, "issue125.txt")); !os.IsNotExist(err) {
		t.Errorf("expected file to be deleted, stat err = %v", err)
	}
}

// The headline inconsistency from the issue: get_path("~/...").exists was false
// even when full_path pointed at a real file.
func Test_PathTilde_GetPathExists(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	target := filepath.Join(tmpHome, "exists_check.txt")
	if err := os.WriteFile(target, []byte("x"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	script := `
p = get_path("~/exists_check.txt")
print(p.exists)
print(p.full_path)
`
	setupAndRunCode(t, script, "--color=never")
	expected := fmt.Sprintf("true\n%s/exists_check.txt\n", filepath.ToSlash(tmpHome))
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_PathTilde_FindPaths(t *testing.T) {
	tmpHome := t.TempDir()
	t.Setenv("HOME", tmpHome)

	if err := os.WriteFile(filepath.Join(tmpHome, "a.txt"), []byte("a"), 0644); err != nil {
		t.Fatalf("failed to seed file: %v", err)
	}

	script := `
print(find_paths("~"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ \"a.txt\" ]\n")
	assertNoErrors(t)
}
