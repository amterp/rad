package testing

import "testing"

func TestStringArrays_General(t *testing.T) {
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

func TestStringArrays_IsString(t *testing.T) {
	rsl := `
a string[] = ["a", "b", "c"]
print(a + [1])
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/9 on '+': Cannot join two arrays of different types: string[], mixed array\n")
	resetTestState()
}

func TestStringArrays_CanModify(t *testing.T) {
	rsl := `
a string[] = ["a", "b", "c"]
a += ["d"]
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[a, b, c, d]\n")
	assertNoErrors(t)
	resetTestState()
}
