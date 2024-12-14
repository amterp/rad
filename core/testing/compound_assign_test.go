package testing

import "testing"

func TestCompound_IntAssignments(t *testing.T) {
	rsl := `a = 1
a += 3 + 4
print(a)
a -= 3 * 2
print(a)
a *= 3
print(a)
a /= 4
print(a)`

	setupAndRunCode(t, rsl)
	expected := `8
2
6
1
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestCompound_FloatAssignments(t *testing.T) {
	rsl := `b = 1.5
b += 3.3
print(b)
b -= 2
print(b)
b *= 4
print(b)
b /= 2.5
print(b)`

	setupAndRunCode(t, rsl)
	expected := `4.8
2.8
11.2
4.4799999999999995
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestCompound_StringAssignments(t *testing.T) {
	rsl := `c = "hi"
c += " there"
print(c)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hi there\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompound_AddIntArray(t *testing.T) {
	rsl := `a = [1]
a += [2]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1, 2]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompound_AddFloatArray(t *testing.T) {
	rsl := `a = [1.1]
a += [2.2]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1.1, 2.2]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompound_AddStringArray(t *testing.T) {
	rsl := `a = ["alice"]
a += ["bob"]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[alice, bob]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompound_SubtractFromArrayErrors(t *testing.T) {
	rsl := `a = [1]
a -= 2`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/4 on '-=': Invalid binary operator for mixed array, int\n")
	resetTestState()
}

func TestCompound_DivideFromArrayErrors(t *testing.T) {
	rsl := `a = [1]
a /= 2`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/4 on '/=': Invalid binary operator for mixed array, int\n")
	resetTestState()
}

func TestCompound_MultiplyFromArrayErrors(t *testing.T) {
	rsl := `a = [1]
a *= 2`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/4 on '*=': Invalid binary operator for mixed array, int\n")
	resetTestState()
}

func TestCompound_ErrorsIfAppendNotArray(t *testing.T) {
	rsl := `a = [1]
a += 2`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/4 on '+=': Invalid binary operator for mixed array, int\n")
	resetTestState()
}

func TestCompound_AddThroughCollection(t *testing.T) {
	rsl := `a = [1, 2]
a[1] += 2
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1, 4]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompound_AddThroughNestedCollection(t *testing.T) {
	rsl := `a = { "alice": [1, 2], "bob": [3, 4] }
a["alice"][0] += 2
a.bob[1] += 2
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "{ alice: [3, 2], bob: [3, 6] }\n")
	assertNoErrors(t)
	resetTestState()
}
