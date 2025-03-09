package testing

import "testing"

func Test_Func_Reverse_NoChars(t *testing.T) {
	rsl := `
a = "hello"
print(reverse(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "olleh\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Reverse_With_Nums(t *testing.T) {
	rsl := `
a = 123
print(reverse(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L3:15

  print(reverse(a))
                ^ Got "int" as the 1st argument of reverse(), but must be: string
`
	assertError(t, 1, expected)
	resetTestState()
}
