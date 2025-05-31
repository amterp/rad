package testing

import "testing"

func Test_Str_Escaping_DoesNotEscapeOtherDelimitersInContent_SingleQuoteDoubleQuote(t *testing.T) {
	script := `
print('"hi"')
print('\"hi\"')
`
	setupAndRunCode(t, script, "--color=never")
	expected := `"hi"
\"hi\"
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Str_Escaping_DoesNotEscapeOtherDelimitersInContent_DoubleQuoteSingleQuote(t *testing.T) {
	script := `
print("'hi'")
print("\'hi\'")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `'hi'
\'hi\'
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Str_Escaping_DoesNotEscapeOtherDelimitersInContent_BacktickSingleQuote(t *testing.T) {
	script := "print('`hi`')\n"
	script += "print('\\`hi\\`')\n"
	setupAndRunCode(t, script, "--color=never")
	expected := "`hi`\n"
	expected += "\\`hi\\`\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
