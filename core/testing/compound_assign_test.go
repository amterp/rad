package testing

import "testing"

func TestCompoundIntAssignments(t *testing.T) {
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

func TestCompoundFloatAssignments(t *testing.T) {
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

func TestCompoundStringAssignments(t *testing.T) {
	rsl := `c = "hi"
c += " there"
print(c)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "hi there\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompoundAddIntArray(t *testing.T) {
	rsl := `a = [1]
a += 2
a += [3]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1, 2, 3]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompoundAddFloatArray(t *testing.T) {
	rsl := `a = [1.1]
a += 2.2
a += [3.3]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[1.1, 2.2, 3.3]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompoundAddStringArray(t *testing.T) {
	rsl := `a = ["alice"]
a += "bob"
a += ["charlie"]
print(a)`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "[alice, bob, charlie]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestCompoundSubtractFromArrayErrors(t *testing.T) {
	rsl := `a = [1]
a -= 2`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/4 on '-=': Invalid binary operator for mixed array, int\n")
	resetTestState()
}

func TestCompoundDivideFromArrayErrors(t *testing.T) {
	rsl := `a = [1]
a /= 2`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/4 on '/=': Invalid binary operator for mixed array, int\n")
	resetTestState()
}

func TestCompoundMultiplyFromArrayErrors(t *testing.T) {
	rsl := `a = [1]
a *= 2`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L2/4 on '*=': Invalid binary operator for mixed array, int\n")
	resetTestState()
}
