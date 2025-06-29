package testing

import (
	"os"
	"testing"
)

func Test_Func_ReadFile(t *testing.T) {
	script := `
a = read_file("data/test_file.txt")
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "size_bytes": 9, "content": "hello bob" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadFile_OneArg(t *testing.T) {
	script := `
a = read_file("data/test_file.txt")
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "size_bytes": 9, "content": "hello bob" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadFile_NoExist(t *testing.T) {
	script := `
a = read_file("does_not_exist.txt")
print(b)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  a = read_file("does_not_exist.txt")
      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
      open does_not_exist.txt: no such file or directory (RAD20005)
`
	assertError(t, 1, expected)
}

func Test_Func_ReadFile_NoPermission(t *testing.T) {
	filePath := "data/no_permission.txt"

	info, _ := os.Stat(filePath)
	originalPerms := info.Mode().Perm()

	os.Chmod(filePath, originalPerms&^0444)
	defer os.Chmod(filePath, originalPerms)

	script := `
a = read_file("data/no_permission.txt")
print(b)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  a = read_file("data/no_permission.txt")
      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
      open data/no_permission.txt: permission denied (RAD20004)
`
	assertError(t, 1, expected)
}

func Test_Func_ReadFile_ErrorsOnDirectory(t *testing.T) {
	script := `
a = read_file("data/")
print(b)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  a = read_file("data/")
      ^^^^^^^^^^^^^^^^^^ read data/: is a directory (RAD20003)
`
	assertError(t, 1, expected)
}
