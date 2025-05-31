package testing

import "testing"

func TestLen_Array(t *testing.T) {
	script := `
a = [40, 50, 60]
print(len(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}

func TestLen_String(t *testing.T) {
	script := `
a = "alice"
print(len(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n")
	assertNoErrors(t)
}

func TestLen_EmojiString(t *testing.T) {
	script := `
a = "alice 👋"
print(len(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "7\n")
	assertNoErrors(t)
}

func TestLen_Map(t *testing.T) {
	script := `
a = { "alice": 40, "bob": "bar", "charlie": [1, "hi"] }
print(len(a))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "3\n")
	assertNoErrors(t)
}
