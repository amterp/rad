package testing

import (
	"os"
	"testing"
)

func Test_Func_WriteFile(t *testing.T) {
	filePath := "data/test_write.txt"
	rsl := `
a, b = write_file("data/test_write.txt", "hello world")
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "bytes_written": 11, "path": "data/test_write.txt" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	os.Remove(filePath)
}

func Test_Func_WriteFile_Append(t *testing.T) {
	filePath := "data/test_write_append.txt"
	os.WriteFile(filePath, []byte("hello"), 0644)

	rsl := `
a, b = write_file("data/test_write_append.txt", " world", append=true)
print(a)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "bytes_written": 6, "path": "data/test_write_append.txt" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	os.Remove(filePath)
}

func Test_Func_WriteFile_NoPermission(t *testing.T) {
	filePath := "data/no_permission_write.txt"

	os.WriteFile(filePath, []byte("initial"), 0644)
	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}
	originalPerms := info.Mode().Perm()

	// Remove write permission.
	os.Chmod(filePath, originalPerms&^0222)

	rsl := `
a, b = write_file("data/no_permission_write.txt", "content")
print(b)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "code": "RAD20004", "msg": "open data/no_permission_write.txt: permission denied" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)

	os.Chmod(filePath, originalPerms)
	os.Remove(filePath)
}

func Test_Func_WriteFile_ErrorsOnDirectory(t *testing.T) {
	rsl := `
a, b = write_file("data/", "content")
print(b)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "code": "RAD20006", "msg": "open data/: is a directory" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_WriteFile_ExitErrorsIfNoErrVar(t *testing.T) {
	rsl := `
a = write_file("does_not_exist_dir/test.txt", "content")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:5

  a = write_file("does_not_exist_dir/test.txt", "content")
      ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
      open does_not_exist_dir/test.txt: no such file or directory (error RAD20005)
`
	assertError(t, 1, expected)
}
