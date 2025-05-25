package testing

import (
	"fmt"
	"os"
	"testing"
)

func Test_GetPath_DoesNotExist(t *testing.T) {
	rsl := `
p = "does_not_exist"
path = get_path(p)
print(path.keys())
vals = path.values()
print("/{p}" in str(vals))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "exists", "full_path" ]
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_GetPath_Exists(t *testing.T) {
	rsl := `
p = "data/test_file.txt"
path = get_path(p)
print(path.keys())
print("/{p}" in path.full_path)
print(path.base_name)
print(path.permissions)
print(path.type)
print(path.size_bytes)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "exists", "full_path", "base_name", "permissions", "type", "size_bytes" ]
true
test_file.txt
-rw-r--r--
file
9
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_GetPath_ExpandsHome(t *testing.T) {
	rsl := `
p = "~/.rad"
path = get_path(p)
print(path.full_path)
`
	setupAndRunCode(t, rsl, "--color=never")
	home, _ := os.UserHomeDir()
	expected := fmt.Sprintf("%s/.rad\n", home)
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
