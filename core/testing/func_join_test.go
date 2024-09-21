package testing

import "testing"

func TestStringJoin(t *testing.T) {
	rsl := `
a1 string[] = ["Hi", "there"]
print(join(a1, " "))
print(join(a1, " ", "Alice: "))
print(join(a1, " ", "Alice: ", "!"))
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

func TestIntJoin(t *testing.T) {
	rsl := `
a2 int[] = [1, 2, 3]
print(join(a2, "_"))
print(join(a2, "_", "Nums: "))
print(join(a2, "_", "Nums: ", "_4"))
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

func TestFloatJoin(t *testing.T) {
	rsl := `
a3 float[] = [1.1, 1.2, 1.3]
print(join(a3, " yes "))
print(join(a3, " yes ", "Floats: "))
print(join(a3, " yes ", "Floats: ", " :D"))
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
