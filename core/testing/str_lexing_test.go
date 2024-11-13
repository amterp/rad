package testing

import "testing"

func TestStrLexing_Newline(t *testing.T) {
	rsl := `
print("Hi\nAlice")
print("Hi\\nAlice")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hi\nAlice\nHi\\nAlice\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_NewlineBacktick(t *testing.T) {
	rsl := "print(`Hi\\nAlice`)"
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hi\\nAlice\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_Tab(t *testing.T) {
	rsl := `
print("a\tb")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "a\tb\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_TabBacktick(t *testing.T) {
	rsl := "print(`a\\tb`)"
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "a\\tb\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_EscapeBracket(t *testing.T) {
	rsl := `
print("{upper('alice')}")
print("\{upper('alice')}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n{upper('alice')}\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_EscapeBracketBacktick(t *testing.T) {
	rsl := "print(`{upper('alice')}`)\nprint(`\\{upper('alice')}`)"
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n{upper('alice')}\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_Quotes(t *testing.T) {
	rsl := `
print('single\'quote')
print("single'quote")
print('double"quote')
print("double\"quote")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "single'quote\nsingle'quote\ndouble\"quote\ndouble\"quote\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_Empty(t *testing.T) {
	rsl := `
print("")
print('')
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "\n\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_EmptyBacktick(t *testing.T) {
	rsl := "print(``)"
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_SeveralBackslashes(t *testing.T) {
	rsl := `
print("\\\\")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "\\\\\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_Mixed(t *testing.T) {
	rsl := `
print("\"\n\"")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "\"\n\"\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_DoubleInterp(t *testing.T) {
	rsl := `
x = 1
y = 2
print("{x}{y}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_DoubleInterpBacktick(t *testing.T) {
	rsl := "x = 1\ny = 2\nprint(`{x}{y}`)"
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_EscapingBrackets(t *testing.T) {
	rsl := `
x = 1
print("\\{x}")
print("\\\{x}")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "\\1\n\\{x}\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_Mixed2(t *testing.T) {
	rsl := `
x = 1
print("Hello\n{x}\tWorld!")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hello\n1\tWorld!\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_Mixed2Backticks(t *testing.T) {
	rsl := "x = 1\nprint(`Hello\\n{x}\\tWorld!`)"
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "Hello\\n1\\tWorld!\n")
	assertNoErrors(t)
	resetTestState()
}
func TestStrLexing_Misc(t *testing.T) {
	rsl := `
print("\\")
print("\n\n\n")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "\\\n\n\n\n\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_EscapingIrrelevantChars(t *testing.T) {
	rsl := `
print("\x")
print("\k")
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "\\x\n\\k\n")
	assertNoErrors(t)
	resetTestState()
}

func TestStrLexing_EscapingBacktickInBackticks(t *testing.T) {
	rsl := "print(`\\``)"
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "`\n")
	assertNoErrors(t)
	resetTestState()
}
