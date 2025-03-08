package testing

import "testing"

func TestStringDelimiters_DoubleQuote(t *testing.T) {
	rsl := `
greeting = "hi"
print(greeting)
name = "alice"
print(greeting + " " + name)
print("Pi: {1 + 2.14}")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
hi alice
Pi: 3.14
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestStringDelimiters_SingleQuote(t *testing.T) {
	rsl := `
greeting = 'hi'
print(greeting)
name = "alice"
print(greeting + ' ' + name)
print('Pi: {1 + 2.14}')
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
hi alice
Pi: 3.14
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestStringDelimiters_Backtick(t *testing.T) {
	rsl := "greeting = `hi`\n" +
		"print(greeting)\n" +
		"name = `alice`" + "\n" +
		"print(greeting + ` ` + name)\n" +
		"print(`Pi: {1 + 2.14}`)\n"
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hi
hi alice
Pi: 3.14
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
