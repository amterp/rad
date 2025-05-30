package testing

import "testing"

func TestMap_Dot_Access(t *testing.T) {
	script := `
a = {"alice": 1, "bob": { "charlie": 2 }}
print(a.alice)
print(a.alice + a.bob.charlie)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n3\n")
	assertNoErrors(t)
}

func TestMap_Dot_CanMixAccess(t *testing.T) {
	script := `
a = {"alice": { "bob": { "charlie": 1 } } }
print(a.alice["bob"].charlie)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func TestMap_Dot_Assign(t *testing.T) {
	script := `
a = { "alice": 1 }
a.alice = 2
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": 2 }\n")
	assertNoErrors(t)
}

func TestMap_Dot_MixedAssign(t *testing.T) {
	script := `
a = {"alice": { "bob": { "charlie": 1 } } }
a.alice["bob"].charlie = 3
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "{ \"alice\": { \"bob\": { \"charlie\": 3 } } }\n")
	assertNoErrors(t)
}
