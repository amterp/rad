package testing

import "testing"

// todo maps
//  - key addition/definition
//  - key deletion
//  - contains key
//  - contains value
//  - iteration (this will require some thinking with i)
//  - keyset
//  - valueset
//  - entryset
//  - pick functions integration

func TestMap_CanDeclare(t *testing.T) {
	rsl := `
a = { "alice": 35, "bob": "bar", "charlie": [1, "hi"] }
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "{ alice: 35, bob: bar, charlie: [1, hi] }\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMap_CanExtract(t *testing.T) {
	rsl := `
a = { "alice": 35, "bob": "bar","charlie": [1, "hi"] }
print(a["charlie"][0] + 1)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "2\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMap_CanDeclareWithExpressions(t *testing.T) {
	rsl := `
foo = "bar"
t = true
f = false
a = { "alice": 30 + 5, "bob": foo, upper("charlie"): [1, t or f] }
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "{ alice: 35, bob: bar, CHARLIE: [1, true] }\n")
	assertNoErrors(t)
	resetTestState()
}

// todo
//func TestMap_CanAddByKey(t *testing.T) {
//	rsl := `
//a = { "alice": 35, "bob": "bar", "charlie": [1, "hi"] }
//a["david"] = 20
//print(a)
//`
//	setupAndRunCode(t, rsl, "--NO-COLOR")
//	assertOnlyOutput(t, stdOutBuffer, "a\nb\nc\n")
//	assertNoErrors(t)
//	resetTestState()
//}
