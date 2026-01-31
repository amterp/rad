package testing

import "testing"

func Test_Matches_CompleteMatch_True(t *testing.T) {
	script := `
print(matches("123", r"\d+"))
print(matches("hello", r"[a-z]+"))
print(matches("Hello", r"[A-Z][a-z]+"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_CompleteMatch_False(t *testing.T) {
	script := `
print(matches("hello123", r"\d+"))
print(matches("Hello", r"[a-z]+"))
print(matches("123abc", r"\d+"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `false
false
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_PartialMatch_True(t *testing.T) {
	script := `
print(matches("hello123", r"\d+", partial=true))
print(matches("abc123def", r"\d+", partial=true))
print(matches("Hello world", r"[A-Z]", partial=true))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_PartialMatch_False(t *testing.T) {
	script := `
print(matches("hello", r"\d+", partial=true))
print(matches("abc", r"[A-Z]", partial=true))
print(matches("", r".", partial=true))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `false
false
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_UFCS_Complete(t *testing.T) {
	script := `
print("123".matches(r"\d+"))
print("hello".matches(r"[a-z]+"))
print("hello123".matches(r"\d+"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_UFCS_Partial(t *testing.T) {
	script := `
print("hello123".matches(r"\d+", partial=true))
print("abc def".matches(r"\s", partial=true))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_ComplexPatterns(t *testing.T) {
	script := `
print(matches("user@example.com", r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}"))
print(matches("not-an-email", r"[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}"))
print(matches("abc123", r"^[a-z]+\d+$"))
print(matches("ABC123", r"^[a-z]+\d+$"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
false
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_EdgeCases(t *testing.T) {
	script := `
print(matches("", r""))
print(matches("", r".*"))
print(matches("a", r""))
print(matches("test", r"test"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
false
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_InvalidRegex(t *testing.T) {
	script := `
result = matches("test", "+")
print("Should not reach this")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20024", "Error compiling regex pattern", "missing argument to repetition operator")
}

func Test_Matches_InvalidRegex_UnclosedGroup(t *testing.T) {
	script := `
result = matches("test", "(abc")
print("Should not reach this")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20024", "Error compiling regex pattern", "missing closing )")
}

func Test_Matches_ReturnsBool(t *testing.T) {
	script := `
result = matches("123", r"\d+")
print(type_of(result))
print(result)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `bool
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_InConditionals(t *testing.T) {
	script := `
text = "hello123"
if text.matches(r"\d+", partial=true):
    print("Contains digits")
else:
    print("No digits")

if text.matches(r"\d+"):
    print("All digits")
else:
    print("Not all digits")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Contains digits
Not all digits
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_SpecialCharacters(t *testing.T) {
	script := `
print(matches("a.b", r"a\.b"))
print(matches("a.b", r"a.b"))
print(matches("a*b", r"a\*b"))
print(matches("a+b", r"a\+b"))
print(matches("a?b", r"a\?b"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
true
true
true
true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Matches_Alternation(t *testing.T) {
	script := `
// Complete match: entire string must be one of the alternatives
print(matches("cat", r"cat|dog"))
print(matches("catfish", r"cat|dog"))     // should be false - not complete match
print(matches("aaa", r"a+|b+"))           // alternation with quantifiers
print(matches("ab", r"a+|b+"))            // should be false - matches neither alternative completely

// Partial match: pattern found anywhere in string  
print(matches("catfish", r"cat|dog", partial=true))   // contains "cat"
print(matches("elephant", r"cat|dog", partial=true))  // contains neither
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
false
true
false
true
false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
