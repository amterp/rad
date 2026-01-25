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
	// Error messages are OS-specific, so check for key parts
	assertErrorContains(t, 1,
		"Error at L2:5",
		"a = read_file(\"does_not_exist.txt\")",
		"open does_not_exist.txt:",
		"(RAD20005)",
	)
}

func Test_Func_ReadFile_NoPermission(t *testing.T) {
	// Windows doesn't have Unix-style permission model, skip this test
	if os.Getenv("GOOS") == "windows" || (os.PathSeparator == '\\') {
		t.Skip("Permission test not applicable on Windows")
	}

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
	// Error messages are OS-specific, so check for key parts
	assertErrorContains(t, 1,
		"Error at L2:5",
		"a = read_file(\"data/\")",
		"(RAD20003)",
	)
}
