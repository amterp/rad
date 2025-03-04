package testing

import "testing"

func Test_Func_Trim_NoChars(t *testing.T) {
	rsl := `
a = "  hello 	"
print(trim(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Trim_OneChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim(a, ","))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "!!,hello,!\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Trim_MultipleChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim(a, "!,"))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "hello\n")
	assertNoErrors(t)
	resetTestState()
}
