package testing

import "testing"

func Test_Replace_Simple(t *testing.T) {
	script := `
print(replace("Hi, Alice", "Alice", "Bob"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hi, Bob\n")
	assertNoErrors(t)
}

func Test_Replace_RegexSimple(t *testing.T) {
	script := `
print(replace("Hi, how are you today?", ",.*", " there!"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hi there!\n")
	assertNoErrors(t)
}

func Test_Replace_RegexCapture(t *testing.T) {
	script := `
print(replace("Hi, Charlie Brown", "(.*) Brown", "$1 Grey"))

input = "I really like sandwiches and soup and pizza."
print(replace(input, "I really like (.*) and (.*) and (.*)\.", "I HATE $3, $2, and $1!"))

print(replace("Name: abc", "a(b)c", "$1o$1"))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Hi, Charlie Grey
I HATE pizza, soup, and sandwiches!
Name: bob
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Replace_ReturnsString(t *testing.T) {
	script := `
print(replace("Hi", "Hi", "Hello") + "!")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hello!\n")
	assertNoErrors(t)
}
