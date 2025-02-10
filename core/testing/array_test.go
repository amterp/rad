package testing

import "testing"

func TestArray_CanUseVarsInArrays(t *testing.T) {
	rsl := `
a = "a"
b = 1
c = true
print([a, b, c])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `[ "a", 1, true ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestArray_General(t *testing.T) {
	rsl := `
a = [1, 2, 3]
print(a)
print(join(a, "-"))
print(a + ["4"])
print(a + ["4"])
b = ["a", 3, false, 5.5]
print(b)
print(join(b, "-"))
print(b + ["yo"])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `[ 1, 2, 3 ]
1-2-3
[ 1, 2, 3, "4" ]
[ 1, 2, 3, "4" ]
[ "a", 3, false, 5.5 ]
a-3-false-5.5
[ "a", 3, false, 5.5, "yo" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestArray_NestedArrays(t *testing.T) {
	rsl := `
a = [1, [2, 3], 4]
for b in a:
	print(b)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `1
[ 2, 3 ]
4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

// TODO
func Test_VarPath_Parenthesized(t *testing.T) {
	t.Skip("TODO varpath requires identifier but below is parenthesized func, needs TS  work")
	rsl := `
a = [1, [2, [3, ["four"]], 5]]
print((a[1])[0]) // 2
print((a[1][1])[0]) // 3
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `2
3
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestArray_DeepNesting(t *testing.T) {
	rsl := `
a = [1, [2, [3, ["four"]], 5]]
print(a[0]) // 1
print(a[1]) // [2, [3, [four]], 5]
print(a[1][1]) // [3, [four]]
print(a[1][1][1]) // [four]
print(a[1][1][1][0]) // four
print(a[1][2]) // 5
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `1
[ 2, [ 3, [ "four" ] ], 5 ]
[ 3, [ "four" ] ]
[ "four" ]
four
5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestArray_CanModify(t *testing.T) {
	rsl := `
a = [1, [2, 3], 4]
a += [5.1, "six"]
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, [ 2, 3 ], 4, 5.1, \"six\" ]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestArray_ConcatDoesNotModifyInPlace(t *testing.T) {
	rsl := `
a = [1, 2, 3]
b = a + [4]
print(a)
print(b)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ 1, 2, 3 ]\n[ 1, 2, 3, 4 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestArray_EntryAssignment(t *testing.T) {
	rsl := `
a = [1, 2, "three"]
a[0] = 5
a[1] = false
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ 5, false, \"three\" ]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestArray_EntryCompoundAssignment(t *testing.T) {
	rsl := `
a = [100, 200, 300, 400]
a[0] += 20
a[1] -= 20
a[2] *= 2
a[3] /= 2
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ 120, 180, 600, 200 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestArray_EntryAssignmentOutOfBounds(t *testing.T) {
	rsl := `
a = [100, 200, 300, 400]
a[4] = 500
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L3:3

  a[4] = 500
    ^ Index out of bounds: 4 (length 4)
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestArray_PositiveIndexing(t *testing.T) {
	rsl := `
a = [100, 200, 300, 400]
print(a[0])
print(a[1])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "100\n200\n")
	assertNoErrors(t)
	resetTestState()
}

func TestArray_NegativeIndexing(t *testing.T) {
	rsl := `
a = [100, 200, 300, 400]
print(a[-1])
print(a[-2])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "400\n300\n")
	assertNoErrors(t)
	resetTestState()
}

func TestArray_NegativeIndexAssignment(t *testing.T) {
	rsl := `
a = [100, 200, 300, 400]
a[-1] = 5
a[-2] = 4
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "[ 100, 200, 4, 5 ]\n")
	assertNoErrors(t)
	resetTestState()
}

func TestArray_TooNegativeIndexingGivesError(t *testing.T) {
	rsl := `
a = [100, 200, 300, 400]
print(a[-99])
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	expected := `Error at L3:9

  print(a[-99])
          ^^^ Index out of bounds: -99 (length 4)
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestArray_TooNegativeIndexAssignmentGivesError(t *testing.T) {
	rsl := `
a = [100, 200, 300, 400]
a[-99] = 5
`
	expected := `Error at L3:3

  a[-99] = 5
    ^^^ Index out of bounds: -99 (length 4)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertError(t, 1, expected)
	resetTestState()
}
