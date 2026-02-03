package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func Test_GetPath_Exists(t *testing.T) {
	script := `
p = "data/test_file.txt"
path = get_path(p)
print("/{p}" in path.full_path)
print(path.base_name)
print(path.permissions)
print(path.type)
print(path.size_bytes)
print("modified_millis" in path.keys())
`
	setupAndRunCode(t, script, "--color=never")
	// Windows doesn't have Unix-style permissions, so permissions differ
	expectedPerms := "-rw-r--r--"
	if runtime.GOOS == "windows" {
		expectedPerms = "-rw-rw-rw-"
	}
	expected := fmt.Sprintf(`true
test_file.txt
%s
file
9
true
`, expectedPerms)
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_GetPath_ExpandsHome(t *testing.T) {
	script := `
p = "~/.rad"
path = get_path(p)
print(path.full_path)
`
	setupAndRunCode(t, script, "--color=never")
	home, _ := os.UserHomeDir()
	// Rad normalizes paths to use forward slashes on all platforms
	expected := fmt.Sprintf("%s/.rad\n", filepath.ToSlash(home))
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
