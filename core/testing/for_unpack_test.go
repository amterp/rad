package testing

import "testing"

func Test_For_Unpack_Basic(t *testing.T) {
	rsl := `
for idx, valA, valB in [["a", 10], ["b", 20], ["c", 30]]:
	print(idx, valA, valB)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `0 a 10
1 b 20
2 c 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_Zip(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
b = [10, 20, 30]
for idx, valA, valB in zip(a, b):
	print(idx, valA, valB)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `0 a 10
1 b 20
2 c 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_Four(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
b = [10, 20, 30]
c = ["x", "y", "z"]
d = [100, 200, 300]
for idx, valA, valB, valC, valD in zip(a, b, c, d):
	print(idx, valA, valB, valC, valD)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `0 a 10 x 100
1 b 20 y 200
2 c 30 z 300
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_DoesNotUnpackIfNotEnoughArgs(t *testing.T) {
	rsl := `
for idx, valA in [["a", 10], ["b", 20], ["c", 30]]:
	print(idx, valA)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `0 [ "a", 10 ]
1 [ "b", 20 ]
2 [ "c", 30 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// not 100% sure about the idx part
func Test_For_Unpack_Comprehension(t *testing.T) {
	rsl := `
a = [3, 4, 5]
b = [10, 20, 30]
c = [x * y for _, x, y in zip(a, b)]
print(c)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ 30, 80, 150 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_Unpack_ErrorsOnNonListOfLists(t *testing.T) {
	rsl := `
a = [3, 4, 5]
for idx, valA, valB in a:
	print(idx, valA, valB)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:24

  for idx, valA, valB in a:
                         ^ Expected list of lists, got element type "int"
`
	assertError(t, 1, expected)
}

// not really an assertion of what's desired, but it's expected atm
func Test_For_Unpack_ErrorsAfterOneLoopForInconsistentRight(t *testing.T) {
	rsl := `
a = [ [10, 20], 30 ]
for idx, valA, valB in a:
	print(idx, valA, valB)
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOutput(t, stdOutBuffer, "0 10 20\n")
	expected := `Error at L3:24

  for idx, valA, valB in a:
                         ^ Expected list of lists, got element type "int"
`
	assertError(t, 1, expected)
}

func Test_For_Unpack_ErrorsIfNotEnoughValuesToUnpack(t *testing.T) {
	rsl := `
for idx, valA, valB, valC in [[10, 20], [30, 40]]:
	print(idx, valA, valB, valC)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:30

  for idx, valA, valB, valC in [[10, 20], [30, 40]]:
                               ^^^^^^^^^^^^^^^^^^^^
                               Expected at least 3 values in inner list, got 2
`
	assertError(t, 1, expected)
}

func Test_For_Unpack_CanUnpackEvenIfNotEnoughLefts(t *testing.T) {
	rsl := `
for idx, valA, valB in [["a", 10, 100], ["b", 20, 200], ["c", 30, 300]]:
	print(idx, valA, valB)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `0 a 10
1 b 20
2 c 30
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
