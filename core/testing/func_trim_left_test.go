package testing

import "testing"

func Test_Func_Trim_Left_SingleChar(t *testing.T) {
	script := `
print(trim_left("aaabbb", "a"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "bbb\n") // ALL "a"s removed
	assertNoErrors(t)
}

func Test_Func_Trim_Left_MultiChar(t *testing.T) {
	script := `
print(trim_left(",,!!,hello,!,", "!,"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello,!,\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Left_NoArg_Whitespace(t *testing.T) {
	script := `
print(trim_left("  hello  "))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello  \n")
	assertNoErrors(t)
}
