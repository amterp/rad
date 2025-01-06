package testing

import "testing"

func Test_For_BasicLoop(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
for item in a:
	print(item)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "a\nb\nc\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_For_ILoop(t *testing.T) {
	rsl := `
a = ["a", "b", "c"]
for idx, item in a:
	print(idx, item)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "0 a\n1 b\n2 c\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_For_ChangesInsideAreRemembered(t *testing.T) {
	rsl := `
num = 0
a = ["a", "b", "c"]
for idx, item in a:
	num += idx
print(num)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_For_MapKeyLoop(t *testing.T) {
	rsl := `
a = { "a": 1, "b": 2, "c": 3 }
for key in a:
	print(key)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "a\nb\nc\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_For_MapKeyValueLoop(t *testing.T) {
	rsl := `
a = { "a": 1, "b": 2, "c": 3 }
for key, value in a:
	print(key)
	print(value)
`
	setupAndRunCode(t, rsl)
	assertOnlyOutput(t, stdOutBuffer, "a\n1\nb\n2\nc\n3\n")
	assertNoErrors(t)
	resetTestState()
}

func Test_For_CanLoopThroughString(t *testing.T) {
	rsl := `
a = "hello 👋"
for char in a:
	print(char)
`
	setupAndRunCode(t, rsl)
	expected := `h
e
l
l
o
 
👋
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

// todo RAD-95 below code should be doable via interpolation
func Test_For_CanLoopThroughColoredString(t *testing.T) {
	rsl := `
a = 'h' + blue("el") + 'lo'
for char in a:
	print(char)
`
	setupAndRunCode(t, rsl)
	expected := "h\n"
	expected += blue("e") + "\n"
	expected += blue("l") + "\n"
	expected += "l\n"
	expected += "o\n"
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
