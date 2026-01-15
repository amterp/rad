package testing

import "testing"

func Test_For_Unpack_Basic(t *testing.T) {
	script := `
for valA, valB in [["a", 10], ["b", 20], ["c", 30]] with loop:
	print(loop.idx, valA, valB)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0 a 10
1 b 20
2 c 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_Zip(t *testing.T) {
	script := `
a = ["a", "b", "c"]
b = [10, 20, 30]
for valA, valB in zip(a, b) with loop:
	print(loop.idx, valA, valB)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0 a 10
1 b 20
2 c 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_Four(t *testing.T) {
	script := `
a = ["a", "b", "c"]
b = [10, 20, 30]
c = ["x", "y", "z"]
d = [100, 200, 300]
for valA, valB, valC, valD in zip(a, b, c, d) with loop:
	print(loop.idx, valA, valB, valC, valD)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0 a 10 x 100
1 b 20 y 200
2 c 30 z 300
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_TwoVariables(t *testing.T) {
	// Two variables now means unpack 2 values, not index + item
	script := `
for valA, valB in [["a", 10], ["b", 20], ["c", 30]]:
	print(valA, valB)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `a 10
b 20
c 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_Comprehension(t *testing.T) {
	script := `
a = [3, 4, 5]
b = [10, 20, 30]
c = [x * y for x, y in zip(a, b)]
print(c)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 30, 80, 150 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_ErrorsOnNonListOfLists(t *testing.T) {
	script := `
a = [3, 4, 5]
for valA, valB in a:
	print(valA, valB)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:19

  for valA, valB in a:
                    ^ Cannot unpack "int" into 2 values
`
	assertError(t, 1, expected)
}

func Test_For_Unpack_ErrorsAfterOneLoopForInconsistentRight(t *testing.T) {
	script := `
a = [ [10, 20], 30 ]
for valA, valB in a:
	print(valA, valB)
`
	setupAndRunCode(t, script, "--color=never")
	assertOutput(t, stdOutBuffer, "10 20\n")
	expected := `Error at L3:19

  for valA, valB in a:
                    ^ Cannot unpack "int" into 2 values
`
	assertError(t, 1, expected)
}

func Test_For_Unpack_ErrorsIfNotEnoughValuesToUnpack(t *testing.T) {
	script := `
for valA, valB, valC in [[10, 20], [30, 40]]:
	print(valA, valB, valC)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:25

  for valA, valB, valC in [[10, 20], [30, 40]]:
                          ^^^^^^^^^^^^^^^^^^^^
                          Expected at least 3 values in inner list, got 2
`
	assertError(t, 1, expected)
}

func Test_For_Unpack_CanUnpackEvenIfNotEnoughLefts(t *testing.T) {
	// Extra values in inner list are ignored
	script := `
for valA, valB in [["a", 10, 100], ["b", 20, 200], ["c", 30, 300]]:
	print(valA, valB)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `a 10
b 20
c 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_MigrationErrorHint(t *testing.T) {
	// When user uses old syntax with 'idx' as first variable, show helpful hint
	script := `
for idx, item in [1, 2, 3]:
	print(idx, item)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:18

  for idx, item in [1, 2, 3]:
                   ^^^^^^^^^
                   Cannot unpack "int" into 2 values

Note: The for-loop syntax changed. It looks like you may be using the old syntax.
Old: for idx, item in items:
New: for item in items with loop:
         print(loop.idx, item)
`
	assertError(t, 1, expected)
}

func Test_For_Unpack_MigrationErrorHintThreeVars(t *testing.T) {
	// Migration hint also works for 3+ variables
	script := `
for idx, item, extra in [1, 2, 3]:
	print(idx, item, extra)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:25

  for idx, item, extra in [1, 2, 3]:
                          ^^^^^^^^^
                          Cannot unpack "int" into 3 values

Note: The for-loop syntax changed. It looks like you may be using the old syntax.
Old: for idx, item in items:
New: for item in items with loop:
         print(loop.idx, item)
`
	assertError(t, 1, expected)
}
