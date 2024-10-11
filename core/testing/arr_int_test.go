package testing

import "testing"

func TestIntArrays_General(t *testing.T) {
	rsl := `
a int[] = [1, 2, 3]
print(a)
print(join(a, "-"))
print(a + [4])
print(a + 4)
`
	setupAndRunCode(t, rsl)
	expected := `[1, 2, 3]
1-2-3
[1, 2, 3, 4]
[1, 2, 3, 4]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestIntArrays_IsInt(t *testing.T) {
	rsl := `
a int[] = [1, 2, 3]
print(a + ["4"])
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/9 on '+': Cannot join two arrays of different types: int[], mixed array\n")
	resetTestState()
}

func TestIntArrays_CanModify(t *testing.T) {
	rsl := `
a int[] = [1, 2, 3]
a += [4]
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1, 2, 3, 4]\n")
	assertNoErrors(t)
	resetTestState()
}
