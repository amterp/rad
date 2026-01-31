package testing

import (
	"os"
	"testing"
)

func Test_Func_WriteFile(t *testing.T) {
	filePath := "data/test_write.txt"
	script := `
a, b = write_file("data/test_write.txt", "hello world")
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "bytes_written": 11, "path": "data/test_write.txt" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	os.Remove(filePath)
}

func Test_Func_WriteFile_Append(t *testing.T) {
	filePath := "data/test_write_append.txt"
	os.WriteFile(filePath, []byte("hello"), 0644)

	script := `
a, b = write_file("data/test_write_append.txt", " world", append=true)
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "bytes_written": 6, "path": "data/test_write_append.txt" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	os.Remove(filePath)
}

func Test_Func_WriteFile_NoPermission(t *testing.T) {
	// Windows permission model is different, skip this test on Windows
	if os.PathSeparator == '\\' {
		t.Skip("Permission test not applicable on Windows")
	}

	filePath := "data/no_permission_write.txt"

	os.WriteFile(filePath, []byte("initial"), 0644)
	defer os.Remove(filePath)

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	originalPerms := info.Mode().Perm()

	// Remove write permission.
	os.Chmod(filePath, originalPerms&^0222)
	defer os.Chmod(filePath, originalPerms)

	script := `
a = write_file("data/no_permission_write.txt", "content")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20004", "permission denied")
}

func Test_Func_WriteFile_ErrorsOnDirectory(t *testing.T) {
	script := `
a = write_file("data/", "content")
`
	setupAndRunCode(t, script, "--color=never")
	// Error messages are OS-specific, so check for key parts
	assertErrorContains(t, 1, "RAD20006", "is a directory")
}

func Test_Func_WriteFile_ExitErrorsIfNoErrVar(t *testing.T) {
	script := `
a = write_file("does_not_exist_dir/test.txt", "content")
`
	setupAndRunCode(t, script, "--color=never")
	// Error messages are OS-specific, so check for key parts
	assertErrorContains(t, 1, "RAD20005", "open does_not_exist_dir/test.txt:")
}
