package testing

import "testing"

func Test_StrLexing_Newline(t *testing.T) {
	script := `
print("Hi\nAlice")
print("Hi\\nAlice")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hi\nAlice\nHi\\nAlice\n")
	assertNoErrors(t)
}

func Test_StrLexing_NewlineBacktick(t *testing.T) {
	script := "print(`Hi\\nAlice`)"
	script += "\nprint(`Hi\\\\nAlice`)"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hi\nAlice\nHi\\nAlice\n")
	assertNoErrors(t)
}

func Test_StrLexing_Tab(t *testing.T) {
	script := `
print("a\tb")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\tb\n")
	assertNoErrors(t)
}

func Test_StrLexing_TabBacktick(t *testing.T) {
	script := "print(`a\\tb`)"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\tb\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapeBracket(t *testing.T) {
	script := `
print("{upper('alice')}")
print("\{upper('alice')}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n{upper('alice')}\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapeBracketBacktick(t *testing.T) {
	script := "print(`{upper('alice')}`)\nprint(`\\{upper('alice')}`)"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n{upper('alice')}\n")
	assertNoErrors(t)
}

func Test_StrLexing_Quotes(t *testing.T) {
	script := `
print('single\'quote')
print("single'quote")
print('double"quote')
print("double\"quote")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "single'quote\nsingle'quote\ndouble\"quote\ndouble\"quote\n")
	assertNoErrors(t)
}

func Test_StrLexing_Empty(t *testing.T) {
	script := `
print("")
print('')
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n\n")
	assertNoErrors(t)
}

func Test_StrLexing_EmptyBacktick(t *testing.T) {
	script := "print(``)"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n")
	assertNoErrors(t)
}

func Test_StrLexing_SeveralBackslashes(t *testing.T) {
	script := `
print("\\\\")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\\\\n")
	assertNoErrors(t)
}

func Test_StrLexing_Mixed(t *testing.T) {
	script := `
print("\"\n\"")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\"\n\"\n")
	assertNoErrors(t)
}

func Test_StrLexing_DoubleInterp(t *testing.T) {
	script := `
x = 1
y = 2
print("{x}{y}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
}

func Test_StrLexing_DoubleInterpBacktick(t *testing.T) {
	script := "x = 1\ny = 2\nprint(`{x}{y}`)"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapingBrackets(t *testing.T) {
	script := `
x = 1
print("\\{x}")
print("\\\{x}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\1\n\\{x}\n")
	assertNoErrors(t)
}

func Test_StrLexing_Mixed2(t *testing.T) {
	script := `
x = 1
print("Hello\n{x}\tWorld!")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hello\n1\tWorld!\n")
	assertNoErrors(t)
}

func Test_StrLexing_Mixed2Backticks(t *testing.T) {
	script := "x = 1\nprint(`Hello\\n{x}\\tWorld!`)"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hello\n1\tWorld!\n")
	assertNoErrors(t)
}
func Test_StrLexing_Misc(t *testing.T) {
	script := `
print("\\")
print("\n\n\n")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\\n\n\n\n\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapingIrrelevantCharsPrintsAsIs(t *testing.T) {
	script := `
print("\x")
print("\k")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\x\n\\k\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapingBacktickInBackticks(t *testing.T) {
	script := "print(`\\``)"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "`\n")
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_DoubleQuotes(t *testing.T) {
	script := `
name = "alice"
print("hi\n{name}")
print(r"hi\n{name}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `hi
alice
hi\n{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_SingleQuotes(t *testing.T) {
	script := `
name = 'alice'
print('hi\n{name}')
print(r'hi\n{name}')
`
	setupAndRunCode(t, script, "--color=never")
	expected := `hi
alice
hi\n{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_Backticks(t *testing.T) {
	script := "name = `alice`\n"
	script += "print(`hi\\n{name}`)\n"
	script += "print(r`hi\\n{name}`)\n"
	setupAndRunCode(t, script, "--color=never")
	expected := `hi
alice
hi\n{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_DoubleSlashIsTwoSlashes(t *testing.T) {
	script := `
print(r"\\")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `\\`+"\n")
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_SingleBackslash(t *testing.T) {
	script := `
print(r"\")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `\`+"\n")
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_ErrorsIfTryingToEscapeDelimiter(t *testing.T) {
	script := `
print(r"\"")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:7

  print(r"\"")
        ^^^^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func Test_StrLexing_Multiline_Simple(t *testing.T) {
	script := `
text = """
Hi Alice
How are you?
"""
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Hi Alice
How are you?
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_ExtraStartingAndEndingNewline(t *testing.T) {
	script := `
text = """

Hi Alice
How are you?

"""
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `
Hi Alice
How are you?

`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_Indented(t *testing.T) {
	script := `
text = """
zero
 one
   three
"""
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `zero
 one
   three
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_RemovesPrefixedWhiteSpaceDependingOnEndingDelimiter(t *testing.T) {
	script := `
text = """
  one
   two
     four
 """
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := ` one
  two
    four
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_InterpolationAndSpecials(t *testing.T) {
	script := `
name = "alice"
text = """
  Hi\n
 there\t{name}
 """
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := ` Hi

there	alice
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_Raw(t *testing.T) {
	script := `
name = "alice"
text = r"""
  Hi\n
 there\t{name}
 """
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := ` Hi\n
there\t{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo
func Test_StrLexing_Multiline_CanCollapseLinesWithBackSlash(t *testing.T) {
	t.Skip("Not implemented")
	script := `
name = "alice"
text = """
  Hi\
 {name} \
 how are you?
 """
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := ` Hialice how are you?
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_CanEscapeTripleQuote(t *testing.T) {
	script := `
name = "alice"
text = """
Text1
\"""
Text2
"""
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Text1
"""
Text2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo
func Test_StrLexing_Multiline_CanIncreaseDelimiterSize(t *testing.T) {
	t.Skip("Not implemented")
	script := `
name = "alice"
text = """"
Text1
"""
Text2
""""
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Text1
"""
Text2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_ErrorIfAnythingAfterOpener(t *testing.T) {
	script := `
text = """abc
 Hi Alice
  How are you?
 """
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:8

  text = """abc
         ^^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func Test_StrLexing_Multiline_NoErrorIfCommentFollowsOpener(t *testing.T) {
	script := `
text = """ // test!
 Hi Alice
  How are you?
 """
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Hi Alice
 How are you?
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_CanConcatOnSameLineAsCloser(t *testing.T) {
	script := `
text = """ // test!
 Hi Alice
  How are you?
 """ + " :)"
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Hi Alice
 How are you? :)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_ErrorIfClosingQuoteIsLessThan3(t *testing.T) {
	script := `
text = """
 Hi Alice
  How are you?
 ""
print(text)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  text = """
  ^ Invalid syntax
`
	assertError(t, 1, expected)
}

// todo
// - ${} syntax?
