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
	expected := `Error at L2:16

  print("hello, {var}")
                 ^^^ Undefined variable: var
`
	assertError(t, 1, expected)
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
//func TestStringInterpolation_CanEscapeSecond(t *testing.T) {
//	rl := `
//print("hello, {var\}")
//`
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
