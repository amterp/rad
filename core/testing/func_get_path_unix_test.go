//go:build unix

package testing

import "testing"

func Test_GetPath_AccessedMillis(t *testing.T) {
	script := `
path = get_path("data/test_file.txt")

// accessed_millis should exist on Unix
print("accessed_millis" in path.keys())

atime = path.accessed_millis

// Should be more recent than 2003
print(atime > 1065885429278)

// Should be in the past or present (before year 2100)
print(atime < 4102444800000)  // 2100-01-01 in millis
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_GetPath_AccessedMillis_Dir(t *testing.T) {
	script := `
path = get_path("data")
print("accessed_millis" in path.keys())
print(path.accessed_millis > 0)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
