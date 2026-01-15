package testing

import "testing"

func Test_For_BasicLoop(t *testing.T) {
	script := `
a = ["a", "b", "c"]
for item in a:
	print(item)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\nb\nc\n")
	assertNoErrors(t)
}

func Test_For_ILoop(t *testing.T) {
	script := `
a = ["a", "b", "c"]
for item in a with loop:
	print(loop.idx, item)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "0 a\n1 b\n2 c\n")
	assertNoErrors(t)
}

func Test_For_ChangesInsideAreRemembered(t *testing.T) {
	script := `
num = 0
a = ["a", "b", "c"]
for item in a with loop:
	num += loop.idx
print(num)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func Test_For_MapKeyLoop(t *testing.T) {
	script := `
a = { "a": 1, "b": 2, "c": 3 }
for key in a:
	print(key)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\nb\nc\n")
	assertNoErrors(t)
}

func Test_For_MapKeyValueLoop(t *testing.T) {
	script := `
a = { "a": 1, "b": 2, "c": 3 }
for key, value in a:
	print(key)
	print(value)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\n1\nb\n2\nc\n3\n")
	assertNoErrors(t)
}

func Test_For_CanLoopThroughString(t *testing.T) {
	script := `
a = "hello 👋"
for char in a:
	print(char)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `h
e
l
l
o
 
👋
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo RAD-95 below code should be doable via interpolation
func Test_For_CanLoopThroughColoredString(t *testing.T) {
	script := `
a = 'h' + blue("el") + 'lo'
for char in a:
	print(char)
`
	setupAndRunCode(t, script, "--color=never")
	expected := "h\n"
	expected += blue("e") + "\n"
	expected += blue("l") + "\n"
	expected += "l\n"
	expected += "o\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_CanContinue(t *testing.T) {
	script := `
for i in range(5):
	if i == 2:
		continue
	print(i)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0
1
3
4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_For_MapCanBreak(t *testing.T) {
	script := `
m = {"a": 1, "b": 2, "c": 3}
for key in m:
	print(key)
	break
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "a\n")
	assertNoErrors(t)
}

func Test_For_MapCanBreakFromSwitch(t *testing.T) {
	// Regression test: break inside switch should exit the for loop, not just the switch.
	// The original bug was that the switch statement used 'break' internally, so a user's
	// 'break' would only exit the switch, causing an infinite loop.
	script := `
m = {"a": 1, "b": 2, "c": 3}
for key, val in m:
	switch val:
		case 1:
			print("found one")
			break
		case 2:
			print("found two")
		default:
			print("other")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "found one\n")
	assertNoErrors(t)
}
