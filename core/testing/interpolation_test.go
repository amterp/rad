package testing

import (
	"testing"

	"github.com/amterp/color"
)

func TestStringInterpolation_String(t *testing.T) {
	script := `
var = "alice"
print("hello, {var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, alice\n")
	assertNoErrors(t)
}

func TestStringInterpolation_Int(t *testing.T) {
	script := `
var = 42
print("hello, {var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, 42\n")
	assertNoErrors(t)
}

func TestStringInterpolation_Float(t *testing.T) {
	script := `
var = 12.5
print("hello, {var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, 12.5\n")
	assertNoErrors(t)
}

func TestStringInterpolation_Bool(t *testing.T) {
	script := `
var = true
print("hello, {var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, true\n")
	assertNoErrors(t)
}

func TestStringInterpolation_List(t *testing.T) {
	script := `
var = ["alice", 42]
print("hello, {var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, [ \"alice\", 42 ]\n")
	assertNoErrors(t)
}

func TestStringInterpolation_Map(t *testing.T) {
	script := `
var = { "name": "alice", "age": 42 }
print("hello, {var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, { \"name\": \"alice\", \"age\": 42 }\n")
	assertNoErrors(t)
}

// todo a better error would be to include the whole string e.g.
// "RslError at L2/20 on '\"hello, {var}\"': Undefined variable referenced: var\n"
func TestStringInterpolation_ErrorsIfUnknownVariable(t *testing.T) {
	script := `
print("hello, {var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20028", "Undefined variable: var")
}

func TestStringInterpolation_CanEscapeFirst(t *testing.T) {
	script := `
print("hello, \{var}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, {var}\n")
	assertNoErrors(t)
}

// todo this should fail
// func TestStringInterpolation_CanEscapeSecond(t *testing.T) {
//	rl := `
// print("hello, {var\}")
// `
//	setupAndRunCode(t, rl, "--color=never")
//	assertOnlyOutput(t, stdOutBuffer, "hello, {var}\n")
//	assertNoErrors(t)
//	//}

func TestStringInterpolation_Expressions(t *testing.T) {
	script := `
print("hello, {2 + 2}")
a = 2
b = 3
print("hello, {a + b}")
name = "alice"
print("hello, {len(name)}")
print("hello, {len('bob')}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "hello, 4\nhello, 5\nhello, 5\nhello, 3\n")
	assertNoErrors(t)
}

func TestStringInterpolation_FormattingString(t *testing.T) {
	script := `
var = "alice"
print("_{var}_")
print("_{var:16}_")
print("_{var:<16}_")
print("_{var:>16}_")
`
	expected := `_alice_
_           alice_
_alice           _
_           alice_
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_FormattingInt(t *testing.T) {
	script := `
num = 12
print("_{num:.2}_")
print("_{num:16}_")
print("_{num:<16}_")
print("_{num:>16}_")
print("_{num:<16.2}_")
print("_{num:>16.2}_")
print("_{num:.10}_")
`
	expected := `_12.00_
_              12_
_12              _
_              12_
_12.00           _
_           12.00_
_12.0000000000_
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_FormattingFloat(t *testing.T) {
	script := `
pi = 3.14159
print("_{pi:.2}_")
print("_{pi:16}_")
print("_{pi:<16}_")
print("_{pi:>16}_")
print("_{pi:<16.2}_")
print("_{pi:>16.2}_")
print("_{pi:.10}_")
`
	expected := `_3.14_
_        3.141590_
_3.141590        _
_        3.141590_
_3.14            _
_            3.14_
_3.1415900000_
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_FormattingFloatExpressions(t *testing.T) {
	script := `
print("_{1 + 2.14159:.2}_")
print("_{1 + 2.14159:16}_")
print("_{1 + 2.14159:<16}_")
print("_{1 + 2.14159:>16}_")
print("_{1 + 2.14159:<16.2}_")
print("_{1 + 2.14159:>16.2}_")
print("_{1 + 2.14159:.10}_")
`
	expected := `_3.14_
_        3.141590_
_3.141590        _
_        3.141590_
_3.14            _
_            3.14_
_3.1415900000_
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_Formatting_ColorDoesNotImpactPadding(t *testing.T) {
	// for some reason, the 'shared blue' has nocolor=true when this test is
	// run by itself, so it fails.... no clue why
	myBlue := color.New(color.FgBlue)
	myBlue.EnableColor()

	script := `
n = "alice"
print("{n:20}")
print("{blue(n):20}")
print("{n:<20}")
print("{blue(n):<20}")
`
	expected := "               alice\n"
	expected += "               " + myBlue.Sprintf("alice") + "\n"
	expected += "alice               \n"
	expected += myBlue.Sprintf("alice") + "               \n"
	setupAndRunCode(t, script, "--color=always")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_Formatting_ColorWorksWithoutPadding(t *testing.T) {
	// for some reason, the 'shared blue' has nocolor=true when this test is
	// run by itself, so it fails.... no clue why
	myBlue := color.New(color.FgBlue)
	myBlue.EnableColor()

	script := `
n = "alice"
print("{n}")
print("{blue(n)}")
`
	expected := "alice\n"
	expected += myBlue.Sprintf("alice") + "\n"
	setupAndRunCode(t, script, "--color=always")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ModuloOperator(t *testing.T) {
	script := `
a = 10
b = 3
print("10 % 3 = {a % b}")
print("Direct: {10 % 3}")
print("Mixed: {a + b} and {a % b}")
`
	expected := "10 % 3 = 1\nDirect: 1\nMixed: 13 and 1\n"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_PercentCharactersInStrings(t *testing.T) {
	script := `
value = 42
print("Percentage: {value}%")
print("Multiple %% signs")
print("Format-like: {value} % complete")
print("URL encoded: hello%20world")
print("Mixed: {value}% done, {100 - value}% remaining")
`
	expected := "Percentage: 42%\nMultiple %% signs\nFormat-like: 42 % complete\nURL encoded: hello%20world\nMixed: 42% done, 58% remaining\n"
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_Int(t *testing.T) {
	script := `
num = 1234567
print("_{num:,}_")
print("_{num:<15,}_")
print("_{num:>15,}_")
print("_{num:15,}_")
`
	expected := `_1,234,567_
_1,234,567      _
_      1,234,567_
_      1,234,567_
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_Float(t *testing.T) {
	script := `
pi = 3141.59265
print("_{pi:,}_")
print("_{pi:,.2}_")
print("_{pi:,.4}_")
print("_{pi:<20,.2}_")
print("_{pi:>20,.2}_")
`
	expected := `_3,141.59265_
_3,141.59_
_3,141.5927_
_3,141.59            _
_            3,141.59_
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_LargeNumbers(t *testing.T) {
	script := `
million = 1000000
billion = 1000000000
trillion = 1000000000000
print("Million: {million:,}")
print("Billion: {billion:,}")
print("Trillion: {trillion:,}")
`
	expected := `Million: 1,000,000
Billion: 1,000,000,000
Trillion: 1,000,000,000,000
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_NegativeNumbers(t *testing.T) {
	script := `
negative_int = -123456
negative_float = -9876.543
print("Negative int: {negative_int:,}")
print("Negative float: {negative_float:,.2}")
print("Padded negative: {negative_int:>15,}")
`
	expected := `Negative int: -123,456
Negative float: -9,876.54
Padded negative:        -123,456
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_EdgeCases(t *testing.T) {
	script := `
zero = 0
small = 123
print("Zero: {zero:,}")
print("Small number: {small:,}")
print("Expression: {1000 + 2000:,}")
`
	expected := `Zero: 0
Small number: 123
Expression: 3,000
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_IntWithPrecision(t *testing.T) {
	script := `
num = 12345
print("Int with precision: {num:,.2}")
print("Int with larger precision: {num:,.6}")
`
	expected := `Int with precision: 12,345.00
Int with larger precision: 12,345.000000
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_TypeSafety(t *testing.T) {
	script := `
name = "John"
print("Testing type safety: {name:,}")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30003", "Cannot format str with thousands separator ','")
}

func TestStringInterpolation_ThousandsSeparator_ScientificNotation(t *testing.T) {
	script := `
big_number = 1e6
very_big = 1.23e9
small_number = 1.23e-3
print("1e6: {big_number:,}")
print("1.23e9: {very_big:,}")
print("1.23e-3: {small_number:,}")
`
	expected := `1e6: 1,000,000
1.23e9: 1,230,000,000
1.23e-3: 0.00123
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_PreservesPrecision(t *testing.T) {
	script := `
tiny_number = 1.23456e-5
print("With comma: {tiny_number:,}")
print("Without comma: {tiny_number}")
`
	expected := `With comma: 0.0000123456
Without comma: 0.0000123456
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestStringInterpolation_ThousandsSeparator_PrecisionConsistency(t *testing.T) {
	script := `
num = 1234.56
print("Without comma: {num}")
print("With comma: {num:,}")
print("Both with precision: {num:.2} vs {num:,.2}")
`
	expected := `Without comma: 1234.56
With comma: 1,234.56
Both with precision: 1234.56 vs 1,234.56
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
