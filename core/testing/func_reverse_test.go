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
                ^
                Value '123' (int) is not compatible with expected type 'str|list'
`
	assertError(t, 1, expected)
}

func Test_Func_Reverse_List_Basic(t *testing.T) {
	script := `
print(reverse([1, 2, 3]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 3, 2, 1 ]\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_List_Empty(t *testing.T) {
	script := `
print(reverse([]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ ]\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_List_SingleElement(t *testing.T) {
	script := `
print(reverse([42]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 42 ]\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_List_MixedTypes(t *testing.T) {
	script := `
print(reverse([1, "a", true]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ true, \"a\", 1 ]\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_List_Strings(t *testing.T) {
	script := `
print(reverse(["hello", "world"]))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ \"world\", \"hello\" ]\n")
	assertNoErrors(t)
}

func Test_Func_Reverse_MultiByte(t *testing.T) {
	script := `
print(reverse("a😀b"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "b😀a\n")
	assertNoErrors(t)
}
