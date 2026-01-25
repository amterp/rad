package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func Test_GetPath_DoesNotExist(t *testing.T) {
	script := `
p = "does_not_exist"
path = get_path(p)
print(path.keys())
vals = path.values()
print("/{p}" in str(vals))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "exists", "full_path" ]
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

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

func Test_GetPath_ModifiedMillis(t *testing.T) {
	script := `
path = get_path("data/test_file.txt")
mtime = path.modified_millis

// Should be newer than 2003
print(mtime > 1065885429278)
// Should be in the past or present (before year 2100)
print(mtime < 4102444800000)  // 2100-01-01 in millis
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_GetPath_ModifiedMillis_Dir(t *testing.T) {
	script := `
path = get_path("data")
print("modified_millis" in path.keys())
print(path.modified_millis > 1065885429278)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
`
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
