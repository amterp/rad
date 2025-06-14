package testing

import "testing"

func Test_Func_Reverse_NoChars(t *testing.T) {
	script := `
a = "hello"
print(reverse(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "olleh\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_EmptyString(t *testing.T) {
	script := `
a = ""
print(reverse(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_StringWithTabSpaces(t *testing.T) {
	script := `
a = "\t\thello\t\n"
print(reverse(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n\tolleh\t\t\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_With_Nums(t *testing.T) {
	script := `
a = 123
print(reverse(a))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:15

  print(reverse(a))
                ^ Got "int" as the 1st argument of reverse(), but must be: str
`
	assertError(t, 1, expected)
}
