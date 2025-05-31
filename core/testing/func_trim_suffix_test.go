package testing

import "testing"

func Test_Func_Trim_Suffix_NoChars(t *testing.T) {
	script := `
a = "	hello	"
print(trim_suffix(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\thello\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Suffix_OneChar(t *testing.T) {
	script := `
a = ",,!!,hello,!,"
print(trim_suffix(a, ","))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ",,!!,hello,!\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Suffix_MultipleChar(t *testing.T) {
	script := `
a = ",,!!,hello,!,"
print(trim_suffix(a, "!,"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ",,!!,hello\n")
	assertNoErrors(t)
}
