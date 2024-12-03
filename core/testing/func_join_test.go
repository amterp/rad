package testing

import "testing"

func TestJoin_String(t *testing.T) {
	rsl := `
arr = ["Hi", "there"]
print(join(arr, " "))
print(join(arr, " ", "Alice: "))
print(join(arr, " ", "Alice: ", "!"))
`
	setupAndRunCode(t, rsl)
	expected := `Hi there
Alice: Hi there
Alice: Hi there!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestJoin_Int(t *testing.T) {
	rsl := `
arr = [1, 2, 3]
print(join(arr, "_"))
print(join(arr, "_", "Nums: "))
print(join(arr, "_", "Nums: ", "_4"))
`
	setupAndRunCode(t, rsl)
	expected := `1_2_3
Nums: 1_2_3
Nums: 1_2_3_4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestJoin_Float(t *testing.T) {
	rsl := `
arr = [1.1, 1.2, 1.3]
print(join(arr, " yes "))
print(join(arr, " yes ", "Floats: "))
print(join(arr, " yes ", "Floats: ", " :D"))
`
	setupAndRunCode(t, rsl)
	expected := `1.1 yes 1.2 yes 1.3
Floats: 1.1 yes 1.2 yes 1.3
Floats: 1.1 yes 1.2 yes 1.3 :D
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestJoin_Mixed(t *testing.T) {
	rsl := `
arr = ["alice", 2]
print(join(arr, "_"))
print(join(arr, "_", "("))
print(join(arr, "_", "(", ")"))
`
	setupAndRunCode(t, rsl)
	expected := `alice_2
(alice_2
(alice_2)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestJoin_ReturnsRslString(t *testing.T) {
	rsl := `
print("Hi " + join(["there", "!"], ""))
`
	setupAndRunCode(t, rsl)
	expected := `Hi there!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
