package testing

import "testing"

func Test_Func_Trim_Prefix_NoChars(t *testing.T) {
	script := `
a = "	hello	"
print(trim_prefix(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello	\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Prefix_OneChar(t *testing.T) {
	script := `
a = ",,!!,hello,!,"
print(trim_prefix(a, ","))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "!!,hello,!,\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Prefix_MultipleChar(t *testing.T) {
	script := `
a = ",,!!,hello,!,"
print(trim_prefix(a, "!,"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello,!,\n")
	assertNoErrors(t)
}
