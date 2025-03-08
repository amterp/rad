package testing

import "testing"

func Test_Func_Trim_Suffix_NoChars(t *testing.T) {
	rsl := `
a = "	hello	"
print(trim_suffix(a))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\thello\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Trim_Suffix_OneChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim_suffix(a, ","))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ",,!!,hello,!\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_Func_Trim_Suffix_MultipleChar(t *testing.T) {
	rsl := `
a = ",,!!,hello,!,"
print(trim_suffix(a, "!,"))
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ",,!!,hello\n")
	assertNoErrors(t)
	resetTestState()
}
