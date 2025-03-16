package testing

import "testing"

func Test_Func_Trim_NoChars(t *testing.T) {
	rsl := `
a = "  hello 	"
print(trim(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}

func Test_Func_Trim_OneChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim(a, ","))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "!!,hello,!\n")
	assertNoErrors(t)
}

func Test_Func_Trim_MultipleChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim(a, "!,"))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}
