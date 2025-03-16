package testing

import (
	"os"
	"testing"
)

func Test_Func_ReadFile(t *testing.T) {
	rsl := `
a, b = read_file("data/test_file.txt")
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "size_bytes": 9, "content": "hello bob" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadFile_ErrEmptyIfNoError(t *testing.T) {
	rsl := `
a, b = read_file("data/test_file.txt")
print(b)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadFile_OneArg(t *testing.T) {
	rsl := `
a = read_file("data/test_file.txt")
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "size_bytes": 9, "content": "hello bob" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadFile_NoExist(t *testing.T) {
	rsl := `
a, b = read_file("does_not_exist.txt")
print(b)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "code": "RAD20005", "msg": "open does_not_exist.txt: no such file or directory" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadFile_NoPermission(t *testing.T) {
	filePath := "data/no_permission.txt"

	info, _ := os.Stat(filePath)
	originalPerms := info.Mode().Perm()

	os.Chmod(filePath, originalPerms&^0444)

	rsl := `
a, b = read_file("data/no_permission.txt")
print(b)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "code": "RAD20004", "msg": "open data/no_permission.txt: permission denied" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	os.Chmod(filePath, originalPerms)
}

func Test_Func_ReadFile_ErrorsOnDirectory(t *testing.T) {
	rsl := `
a, b = read_file("data/")
print(b)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "code": "RAD20003", "msg": "read data/: is a directory" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadFile_ExitErrorsIfNoErrVar(t *testing.T) {
	rsl := `
a = read_file("does_not_exist.txt")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:5

  a = read_file("does_not_exist.txt")
      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
      open does_not_exist.txt: no such file or directory (error RAD20005)
`
	assertError(t, 1, expected)
}
