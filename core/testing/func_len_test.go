package testing

import "testing"

func TestLen_Array(t *testing.T) {
	rsl := `
a = [40, 50, 60]
print(len(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}

func TestLen_String(t *testing.T) {
	rsl := `
a = "alice"
print(len(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
	resetTestState()
}

func TestLen_EmojiString(t *testing.T) {
	rsl := `
a = "alice 👋"
print(len(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "7\n")
	assertNoErrors(t)
	resetTestState()
}

func TestLen_Map(t *testing.T) {
	rsl := `
a = { "alice": 40, "bob": "bar", "charlie": [1, "hi"] }
print(len(a))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
	resetTestState()
}
