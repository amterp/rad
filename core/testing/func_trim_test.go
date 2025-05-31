package testing

import "testing"

func Test_Func_Trim_NoChars(t *testing.T) {
	script := `
a = "  hello 	"
print(trim(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}

func Test_Func_Trim_OneChar(t *testing.T) {
	script := `
a = ",,!!,hello,!,"
print(trim(a, ","))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "!!,hello,!\n")
	assertNoErrors(t)
}

func Test_Func_Trim_MultipleChar(t *testing.T) {
	script := `
a = ",,!!,hello,!,"
print(trim(a, "!,"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
}
