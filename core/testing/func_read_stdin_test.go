package testing

import (
	"testing"
)

func Test_Func_ReadStdin_Basic(t *testing.T) {
	script := `
content = read_stdin()
print(content)
`
	tp := NewTestParams(script, "--color=never").StdinInput("hello from stdin")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "hello from stdin\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_MultiLine(t *testing.T) {
	script := `
content = read_stdin()
print(content)
`
	tp := NewTestParams(script, "--color=never").StdinInput("line1\nline2\nline3")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "line1\nline2\nline3\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_Empty(t *testing.T) {
	script := `
content = read_stdin()
if content == null:
    print("content: null")
else:
    print("content: '{content}'")
`
	tp := NewTestParams(script, "--color=never").StdinInput("")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "content: ''\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_NoStdin(t *testing.T) {
	script := `
content = read_stdin()
print("content: {content}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "content: null\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_WithSpecialChars(t *testing.T) {
	script := `
content = read_stdin()
print(content)
`
	tp := NewTestParams(script, "--color=never").StdinInput("hello\tworld\n!@#$%^&*()")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "hello\tworld\n!@#$%^&*()\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_CanCheckNull(t *testing.T) {
	script := `
content = read_stdin()
if content == null:
    print("no stdin")
else:
    print("got stdin: {content}")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "no stdin\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_CanCheckNullWithData(t *testing.T) {
	script := `
content = read_stdin()
if content == null:
    print("no stdin")
else:
    print("got stdin: {content}")
`
	tp := NewTestParams(script, "--color=never").StdinInput("data here")
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "got stdin: data here\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_LongInput(t *testing.T) {
	script := `
content = read_stdin()
print(len(content))
`
	longInput := ""
	for i := 0; i < 1000; i++ {
		longInput += "abcdefghij"
	}
	tp := NewTestParams(script, "--color=never").StdinInput(longInput)
	setupAndRun(t, tp)
	assertOnlyOutput(t, stdOutBuffer, "10000\n")
	assertNoErrors(t)
}

func Test_Func_ReadStdin_ProcessLines(t *testing.T) {
	script := `
content = read_stdin()
if content == null:
    exit(1)

lines = split(content, "\n")
for line in lines:
    if line != "":
        print("Line: {line}")
`
	tp := NewTestParams(script, "--color=never").StdinInput("first\nsecond\nthird")
	setupAndRun(t, tp)
	expected := `Line: first
Line: second
Line: third
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_ReadStdin_MultipleReads(t *testing.T) {
	script := `
first = read_stdin()
second = read_stdin()

print("first: '{first}'")
print("second: '{second}'")
print("first is null: {first == null}")
print("second is null: {second == null}")
`
	tp := NewTestParams(script, "--color=never").StdinInput("test data")
	setupAndRun(t, tp)
	expected := `first: 'test data'
second: ''
first is null: false
second is null: false
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
