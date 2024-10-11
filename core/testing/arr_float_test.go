package testing

import "testing"

func TestFloatArrays(t *testing.T) {
	rsl := `
a float[] = [1.1, 2.2, 3.3]
print(a)
print(join(a, "-"))
print(a + [4.4])
print(a + 4.4)
print(a + [4])
print(a + 4)
`
	setupAndRunCode(t, rsl)
	expected := `[1.1, 2.2, 3.3]
1.1-2.2-3.3
[1.1, 2.2, 3.3, 4.4]
[1.1, 2.2, 3.3, 4.4]
[1.1, 2.2, 3.3, 4]
[1.1, 2.2, 3.3, 4]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestFloatArrays_IsFloat(t *testing.T) {
	rsl := `
a float[] = [1.1, 2.2, 3.3]
print(a + ["4.4"])
`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L3/9 on '+': Cannot join two arrays of different types: float[], mixed array\n")
	resetTestState()
}

func TestFloatArrays_CanModify(t *testing.T) {
	rsl := `
a float[] = [1.1, 2.2, 3.3]
a += [4.4]
print(a)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1.1, 2.2, 3.3, 4.4]\n")
	assertNoErrors(t)
	resetTestState()
}
