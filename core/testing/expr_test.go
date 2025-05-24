package testing

import "testing"

func Test_Not(t *testing.T) {
	rsl := `
a = false
if not a:
    print("it works!")
if not not not a:
    print("it works!!!")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `it works!
it works!!!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_ExprThenIndexing(t *testing.T) {
	rsl := `
a = [4, 2, 3, 1]
print(sort(a)[0])
print(sort(a)[2:][-1])
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `1
4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_FalsyCoalescing(t *testing.T) {
	rsl := `
print(0 or 0.0 or "" or [] or {} or null or false or "hello!")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `hello!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_CanMultiplyStrings(t *testing.T) {
	rsl := `
print("hello!" * 3)
print("hello!".red() * 3)
`
	setupAndRunCode(t, rsl, "--color=always")
	expected := "hello!hello!hello!\n\x1b[31mhello!\x1b[0m\x1b[31mhello!\x1b[0m\x1b[31mhello!\x1b[0m\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
