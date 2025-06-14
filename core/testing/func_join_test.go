package testing

import "testing"

func Test_Join_String(t *testing.T) {
	script := `
arr = ["Hi", "there"]
print(join(arr, " "))
print(join(arr, " ", "Alice: "))
print(join(arr, " ", "Alice: ", "!"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Hi there
Alice: Hi there
Alice: Hi there!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Join_Int(t *testing.T) {
	script := `
arr = [1, 2, 3]
print(join(arr, "_"))
print(join(arr, "_", "Nums: "))
print(join(arr, "_", "Nums: ", "_4"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1_2_3
Nums: 1_2_3
Nums: 1_2_3_4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Join_Float(t *testing.T) {
	script := `
arr = [1.1, 1.2, 1.3]
print(join(arr, " yes "))
print(join(arr, " yes ", "Floats: "))
print(join(arr, " yes ", "Floats: ", " :D"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1.1 yes 1.2 yes 1.3
Floats: 1.1 yes 1.2 yes 1.3
Floats: 1.1 yes 1.2 yes 1.3 :D
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Join_Mixed(t *testing.T) {
	script := `
arr = ["alice", 2]
print(join(arr, "_"))
print(join(arr, "_", "("))
print(join(arr, "_", "(", ")"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `alice_2
(alice_2
(alice_2)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Join_ReturnsRadString(t *testing.T) {
	script := `
print(type_of(join(["Hi", "!"], "")))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `str
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
