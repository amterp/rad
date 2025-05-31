package testing

import "testing"

func Test_FindPaths_DefaultTarget(t *testing.T) {
	script := `
paths = find_paths("path_example/dir2")
print(paths)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "dir21", "dir21/file21.txt", "dir22", "dir22/file22.txt", "file2.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FindPaths_Cwd(t *testing.T) {
	script := `
paths = find_paths("path_example/dir2", relative="cwd")
print(paths)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "path_example/dir2/dir21", "path_example/dir2/dir21/file21.txt", "path_example/dir2/dir22", "path_example/dir2/dir22/file22.txt", "path_example/dir2/file2.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FindPaths_Abs(t *testing.T) {
	t.Skip("Need to test abstractions in place for this to work regardless of the computer running the test")
	script := `
paths = find_paths("path_example/dir2", relative="absolute")
print(paths)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FindPaths_Depth(t *testing.T) {
	script := `
paths = find_paths("path_example", depth=1)
print(paths)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "dir1", "dir2" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FindPaths_Neg3DepthIncludesAll(t *testing.T) {
	script := `
paths = find_paths("path_example", depth=-3)
print(paths)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "dir1", "dir1/dir11", "dir1/dir11/file11.txt", "dir2", "dir2/dir21", "dir2/dir21/file21.txt", "dir2/dir22", "dir2/dir22/file22.txt", "dir2/file2.txt" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FindPaths_Depth0IncludesNothing(t *testing.T) {
	script := `
paths = find_paths("path_example", depth=0)
print(paths)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
