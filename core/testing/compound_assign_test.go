package testing

import "testing"

func TestCompound_IntAssignments(t *testing.T) {
	script := `a = 1
a += 3 + 4
print(a)
a -= 3 * 2
print(a)
a *= 3
print(a)
a /= 4
print(a)`

	setupAndRunCode(t, script, "--color=never")
	expected := `8
2
6
1.5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestCompound_FloatAssignments(t *testing.T) {
	script := `b = 1.5
b += 3.3
print(b)
b -= 2
print(b)
b *= 4
print(b)
b /= 2.5
print(b)`

	setupAndRunCode(t, script, "--color=never")
	expected := `4.8
2.8
11.2
4.4799999999999995
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestCompound_StringAssignments(t *testing.T) {
	script := `c = "hi"
c += " there"
print(c)`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hi there\n")
	assertNoErrors(t)
}

func TestCompound_AddIntArray(t *testing.T) {
	script := `a = [1]
a += [2]
print(a)`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, 2 ]\n")
	assertNoErrors(t)
}

func TestCompound_AddFloatArray(t *testing.T) {
	script := `a = [1.1]
a += [2.2]
print(a)`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1.1, 2.2 ]\n")
	assertNoErrors(t)
}

func TestCompound_AddStringArray(t *testing.T) {
	script := `a = ["alice"]
a += ["bob"]
print(a)`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ \"alice\", \"bob\" ]\n")
	assertNoErrors(t)
}

func TestCompound_SubtractFromArrayErrors(t *testing.T) {
	script := `a = [1]
a -= 2`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30002", "Invalid operand types: cannot do 'list -= int'")
}

func TestCompound_DivideFromArrayErrors(t *testing.T) {
	script := `a = [1]
a /= 2`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30002", "Invalid operand types: cannot do 'list /= int'")
}

func TestCompound_MultiplyFromArrayErrors(t *testing.T) {
	script := `a = [1]
a *= 2`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30002", "Invalid operand types: cannot do 'list *= int'")
}

func TestCompound_ErrorsIfAppendNotArray(t *testing.T) {
	script := `a = [1]
a += 2`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30002", "Invalid operand types: cannot do 'list += int'", "Did you mean to wrap the right side in a list in order to append?")
}

func TestCompound_AddThroughCollection(t *testing.T) {
	script := `a = [1, 2]
a[1] += 2
print(a)`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, 4 ]\n")
	assertNoErrors(t)
}

func TestCompound_AddThroughNestedCollection(t *testing.T) {
	script := `a = { "alice": [1, 2], "bob": [3, 4] }
a["alice"][0] += 2
a.bob[1] += 2
print(a)`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": [ 3, 2 ], \"bob\": [ 3, 6 ] }\n")
	assertNoErrors(t)
}
