package testing

import "testing"

func Test_Str_Escaping_DoesNotEscapeOtherDelimitersInContent_SingleQuoteDoubleQuote(t *testing.T) {
	rsl := `
print('"hi"')
print('\"hi\"')
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `"hi"
\"hi\"
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Str_Escaping_DoesNotEscapeOtherDelimitersInContent_DoubleQuoteSingleQuote(t *testing.T) {
	rsl := `
print("'hi'")
print("\'hi\'")
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `'hi'
\'hi\'
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func Test_Str_Escaping_DoesNotEscapeOtherDelimitersInContent_BacktickSingleQuote(t *testing.T) {
	rsl := "print('`hi`')\n"
	rsl += "print('\\`hi\\`')\n"
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := "`hi`\n"
	expected += "\\`hi\\`\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
