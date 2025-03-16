package testing

import "testing"

func Test_StrLexing_Newline(t *testing.T) {
	rsl := `
print("Hi\nAlice")
print("Hi\\nAlice")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hi\nAlice\nHi\\nAlice\n")
	assertNoErrors(t)
}

func Test_StrLexing_NewlineBacktick(t *testing.T) {
	rsl := "print(`Hi\\nAlice`)"
	rsl += "\nprint(`Hi\\\\nAlice`)"
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hi\nAlice\nHi\\nAlice\n")
	assertNoErrors(t)
}

func Test_StrLexing_Tab(t *testing.T) {
	rsl := `
print("a\tb")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\tb\n")
	assertNoErrors(t)
}

func Test_StrLexing_TabBacktick(t *testing.T) {
	rsl := "print(`a\\tb`)"
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\tb\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapeBracket(t *testing.T) {
	rsl := `
print("{upper('alice')}")
print("\{upper('alice')}")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n{upper('alice')}\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapeBracketBacktick(t *testing.T) {
	rsl := "print(`{upper('alice')}`)\nprint(`\\{upper('alice')}`)"
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "ALICE\n{upper('alice')}\n")
	assertNoErrors(t)
}

func Test_StrLexing_Quotes(t *testing.T) {
	rsl := `
print('single\'quote')
print("single'quote")
print('double"quote')
print("double\"quote")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "single'quote\nsingle'quote\ndouble\"quote\ndouble\"quote\n")
	assertNoErrors(t)
}

func Test_StrLexing_Empty(t *testing.T) {
	rsl := `
print("")
print('')
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n\n")
	assertNoErrors(t)
}

func Test_StrLexing_EmptyBacktick(t *testing.T) {
	rsl := "print(``)"
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n")
	assertNoErrors(t)
}

func Test_StrLexing_SeveralBackslashes(t *testing.T) {
	rsl := `
print("\\\\")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\\\\n")
	assertNoErrors(t)
}

func Test_StrLexing_Mixed(t *testing.T) {
	rsl := `
print("\"\n\"")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\"\n\"\n")
	assertNoErrors(t)
}

func Test_StrLexing_DoubleInterp(t *testing.T) {
	rsl := `
x = 1
y = 2
print("{x}{y}")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
}

func Test_StrLexing_DoubleInterpBacktick(t *testing.T) {
	rsl := "x = 1\ny = 2\nprint(`{x}{y}`)"
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "12\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapingBrackets(t *testing.T) {
	rsl := `
x = 1
print("\\{x}")
print("\\\{x}")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\1\n\\{x}\n")
	assertNoErrors(t)
}

func Test_StrLexing_Mixed2(t *testing.T) {
	rsl := `
x = 1
print("Hello\n{x}\tWorld!")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hello\n1\tWorld!\n")
	assertNoErrors(t)
}

func Test_StrLexing_Mixed2Backticks(t *testing.T) {
	rsl := "x = 1\nprint(`Hello\\n{x}\\tWorld!`)"
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hello\n1\tWorld!\n")
	assertNoErrors(t)
}
func Test_StrLexing_Misc(t *testing.T) {
	rsl := `
print("\\")
print("\n\n\n")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\\n\n\n\n\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapingIrrelevantCharsPrintsAsIs(t *testing.T) {
	rsl := `
print("\x")
print("\k")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\\x\n\\k\n")
	assertNoErrors(t)
}

func Test_StrLexing_EscapingBacktickInBackticks(t *testing.T) {
	rsl := "print(`\\``)"
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "`\n")
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_DoubleQuotes(t *testing.T) {
	rsl := `
name = "alice"
print("hi\n{name}")
print(r"hi\n{name}")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
alice
hi\n{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_SingleQuotes(t *testing.T) {
	rsl := `
name = 'alice'
print('hi\n{name}')
print(r'hi\n{name}')
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
alice
hi\n{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_Backticks(t *testing.T) {
	rsl := "name = `alice`\n"
	rsl += "print(`hi\\n{name}`)\n"
	rsl += "print(r`hi\\n{name}`)\n"
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
alice
hi\n{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_DoubleSlashIsTwoSlashes(t *testing.T) {
	rsl := `
print(r"\\")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `\\`+"\n")
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_SingleBackslash(t *testing.T) {
	rsl := `
print(r"\")
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, `\`+"\n")
	assertNoErrors(t)
}

func Test_StrLexing_RawStrings_ErrorsIfTryingToEscapeDelimiter(t *testing.T) {
	rsl := `
print(r"\"")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:7

  print(r"\"")
        ^^^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func Test_StrLexing_Multiline_Simple(t *testing.T) {
	rsl := `
text = """
Hi Alice
How are you?
"""
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Hi Alice
How are you?
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_ExtraStartingAndEndingNewline(t *testing.T) {
	rsl := `
text = """

Hi Alice
How are you?

"""
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `
Hi Alice
How are you?

`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_Indented(t *testing.T) {
	rsl := `
text = """
zero
 one
   three
"""
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `zero
 one
   three
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_RemovesPrefixedWhiteSpaceDependingOnEndingDelimiter(t *testing.T) {
	rsl := `
text = """
  one
   two
     four
 """
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := ` one
  two
    four
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_InterpolationAndSpecials(t *testing.T) {
	rsl := `
name = "alice"
text = """
  Hi\n
 there\t{name}
 """
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := ` Hi

there	alice
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_Raw(t *testing.T) {
	rsl := `
name = "alice"
text = r"""
  Hi\n
 there\t{name}
 """
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := ` Hi\n
there\t{name}
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo
func Test_StrLexing_Multiline_CanCollapseLinesWithBackSlash(t *testing.T) {
	t.Skip("Not implemented")
	rsl := `
name = "alice"
text = """
  Hi\
 {name} \
 how are you?
 """
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := ` Hialice how are you?
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_CanEscapeTripleQuote(t *testing.T) {
	rsl := `
name = "alice"
text = """
Text1
\"""
Text2
"""
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
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
	rsl := `
name = "alice"
text = """"
Text1
"""
Text2
""""
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Text1
"""
Text2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_ErrorIfAnythingAfterOpener(t *testing.T) {
	rsl := `
text = """abc
 Hi Alice
  How are you?
 """
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:8

  text = """abc
         ^^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func Test_StrLexing_Multiline_NoErrorIfCommentFollowsOpener(t *testing.T) {
	rsl := `
text = """ // test!
 Hi Alice
  How are you?
 """
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Hi Alice
 How are you?
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_CanConcatOnSameLineAsCloser(t *testing.T) {
	rsl := `
text = """ // test!
 Hi Alice
  How are you?
 """ + " :)"
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Hi Alice
 How are you? :)
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_StrLexing_Multiline_ErrorIfClosingQuoteIsLessThan3(t *testing.T) {
	rsl := `
text = """
 Hi Alice
  How are you?
 ""
print(text)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L2:1

  text = """
  ^ Invalid syntax
`
	assertError(t, 1, expected)
}

// todo
// - ${} syntax?
