package testing

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func Test_Func_Mkdir_CreatesNestedDirs(t *testing.T) {
	dir := filepath.ToSlash(filepath.Join(t.TempDir(), "a", "b", "c"))
	script := fmt.Sprintf(`
res = mkdir("%s")
print(res.created)
`, dir)
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "true\n")
	assertNoErrors(t)

	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatalf("Expected directory to exist at %s, got err=%v", dir, err)
	}
}

func Test_Func_Mkdir_ExistingDirIsIdempotent(t *testing.T) {
	dir := filepath.ToSlash(t.TempDir())
	script := fmt.Sprintf(`
res = mkdir("%s")
print(res.created)
print(res.path)
`, dir)
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, fmt.Sprintf("false\n%s\n", dir))
	assertNoErrors(t)
}

func Test_Func_Mkdir_ErrorsWhenPathIsFile(t *testing.T) {
	file := filepath.Join(t.TempDir(), "occupied.txt")
	if err := os.WriteFile(file, []byte("hi"), 0644); err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}
	script := fmt.Sprintf(`
res = mkdir("%s") catch:
    print("caught")
`, filepath.ToSlash(file))
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "caught\n")
	assertNoErrors(t)
}
