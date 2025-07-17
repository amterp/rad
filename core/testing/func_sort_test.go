package testing

import "testing"

func Test_Sort_Basic(t *testing.T) {
	script := `
a = [3, 4, 2, 1]
print(sort(a, reverse=false))
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2, 3, 4 ]
[ 3, 4, 2, 1 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Sort_MixedTypes(t *testing.T) {
	script := `
a = [1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [3, 1, 2], 1.5, -1.2]
print(sort(a))
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ false, true, -1.2, 1, 1.5, 2, 2, "a", "a", "b", [ 3, 1, 2 ], { "alice": 1 } ]
[ 1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [ 3, 1, 2 ], 1.5, -1.2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Sort_Reverse_Basic(t *testing.T) {
	script := `
a = [3, 4, 2, 1]
b = true
print(sort(a, reverse=b))
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 4, 3, 2, 1 ]
[ 3, 4, 2, 1 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Sort_Reverse_MixedTypes(t *testing.T) {
	script := `
a = [1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [3, 1, 2], 1.5, -1.2]
print(sort(a, reverse=true))
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ { "alice": 1 }, [ 3, 1, 2 ], "b", "a", "a", 2, 2, 1.5, 1, -1.2, true, false ]
[ 1, "a", 2, "b", true, false, { "alice": 1 }, 2, "a", [ 3, 1, 2 ], 1.5, -1.2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Sort_StringReverse(t *testing.T) {
	script := `
sort("alice", reverse=true).print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `lieca
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Sort_ParallelSort(t *testing.T) {
	script := `
a = [2, 1, 4, 3]
b = ["a", "b", "c", "d"]
c = [true, false, true, false]
A, B, C = sort(a, b, c)
print(A, B, C)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2, 3, 4 ] [ "b", "a", "d", "c" ] [ false, true, false, true ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Sort_ParallelSort_Reversed(t *testing.T) {
	script := `
a = [2, 1, 4, 3]
b = ["a", "b", "c", "d"]
c = [true, false, true, false]
A, B, C = sort(a, b, c, reverse=true)
print(A, B, C)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 4, 3, 2, 1 ] [ "c", "d", "a", "b" ] [ true, false, true, false ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
