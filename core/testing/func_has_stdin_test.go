package testing

import (
	"testing"
)

func Test_Func_HasStdin_True(t *testing.T) {
	script := `
if has_stdin():
    print("has stdin")
else:
    print("no stdin")
`
	tp := NewTestParams(script, "--color=never").StdinInput("some data")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "has stdin\n")
	assertNoErrors(t)
}

func Test_Func_HasStdin_EmptyStdin(t *testing.T) {
	script := `
if has_stdin():
    print("has stdin")
else:
    print("no stdin")
`
	tp := NewTestParams(script, "--color=never").StdinInput("")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "has stdin\n")
	assertNoErrors(t)
}

func Test_Func_HasStdin_CheckBeforeRead(t *testing.T) {
	script := `
if has_stdin():
    content = read_stdin()
    print("Content: {content}")
else:
    print("No stdin to read")
`
	tp := NewTestParams(script, "--color=never").StdinInput("hello")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "Content: hello\n")
	assertNoErrors(t)
}

func Test_Func_HasStdin_Assignment(t *testing.T) {
	script := `
has_data = has_stdin()
print("Has stdin: {has_data}")
`
	tp := NewTestParams(script, "--color=never").StdinInput("data")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "Has stdin: true\n")
	assertNoErrors(t)
}
