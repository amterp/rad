package testing

import "testing"

func Test_Func_Trim_Prefix_NoChars(t *testing.T) {
	rsl := `
a = "	hello	"
print(trim_prefix(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello	\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Trim_Prefix_OneChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim_prefix(a, ","))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "!!,hello,!,\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Trim_Prefix_MultipleChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim_prefix(a, "!,"))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello,!,\n")
	assertNoErrors(t)
	resetTestState()
}
