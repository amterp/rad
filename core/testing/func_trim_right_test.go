package testing

import "testing"

func Test_Func_Trim_Right_SingleChar(t *testing.T) {
	script := `
print(trim_right("aaabbb", "b"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "aaa\n") // ALL "b"s removed
	assertNoErrors(t)
}

func Test_Func_Trim_Right_MultiChar(t *testing.T) {
	script := `
print(trim_right(",,!!,hello,!,", "!,"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ",,!!,hello\n")
	assertNoErrors(t)
}

func Test_Func_Trim_Right_NoArg_Whitespace(t *testing.T) {
	script := `
print(trim_right("  hello  "))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "  hello\n")
	assertNoErrors(t)
}
