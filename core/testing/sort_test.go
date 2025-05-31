package testing

import "testing"

func TestSort_Basic(t *testing.T) {
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

func TestSort_MixedTypes(t *testing.T) {
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

func TestSort_Reverse_Basic(t *testing.T) {
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

func TestSort_Reverse_MixedTypes(t *testing.T) {
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
