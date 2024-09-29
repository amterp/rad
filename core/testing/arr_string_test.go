package testing

import "testing"

func TestStringArrays(t *testing.T) {
	rsl := `
a string[] = ["a", "b", "c"]
print(a)
print(join(a, "-"))
print(a + ["d"])
print(a + "d")
print(a + 1)
`
	setupAndRunCode(t, rsl)
	expected := `[a, b, c]
a-b-c
[a, b, c, d]
[a, b, c, d]
[a, b, c, 1]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestStringArrayIsString(t *testing.T) {
	rsl := `
a string[] = ["a", "b", "c"]
print(a + [1])
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/9 on '+': Cannot join two arrays of different types: string[], mixed array\n")
	resetTestState()
}
