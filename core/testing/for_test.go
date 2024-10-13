package testing

import "testing"

func TestFor_BasicLoop(t *testing.T) {
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

func TestFor_ILoop(t *testing.T) {
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

func TestFor_ChangesInsideAreRemembered(t *testing.T) {
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

func TestFor_MapKeyLoop(t *testing.T) {
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

func TestFor_MapKeyValueLoop(t *testing.T) {
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
