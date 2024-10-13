package testing

import "testing"

// todo maps
//  - contains key
//  - contains value
//  - iteration
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

func TestMap_CanAddByKey(t *testing.T) {
	rsl := `
a = { "alice": 35, "bob": "bar"}
a["charlie"] = 20
a[upper("dave")] = "hi"
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "{ alice: 35, bob: bar, charlie: 20, DAVE: hi }\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMap_CanCompoundAssign(t *testing.T) {
	rsl := `
a = { "alice": 100, "bob": 200, "charlie": 300, "dave": 400 }
a["alice"] += 20
a["bob"] -= 20
a["charlie"] *= 2
a["dave"] /= 2
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertOnlyOutput(t, stdOutBuffer, "{ alice: 120, bob: 180, charlie: 600, dave: 200 }\n")
	assertNoErrors(t)
	resetTestState()
}

func TestMap_CompoundOpOnNonExistentKeyErrors(t *testing.T) {
	rsl := `
a = { "alice": 100, "bob": 200, "charlie": 300, "dave": 400 }
a["eve"] += 20
print(a)
`
	setupAndRunCode(t, rsl, "--NO-COLOR")
	assertError(t, 1, "RslError at L3/11 on '+=': Cannot use compound assignment on non-existing map key \"eve\"\n")
	resetTestState()
}

// todo this needs to work
//func TestMap_CanModifyArrayNestedInMap(t *testing.T) {
//	rsl := `
//a = { "alice": 100, "bob": [10, 20, 30] }
//a["bob"][1] = 200
//a["bob"][2] += 5
//print(a)
//`
//	setupAndRunCode(t, rsl, "--NO-COLOR")
//	assertOnlyOutput(t, stdOutBuffer, "{ alice: 100, bob: [10, 200, 35] }\n")
//	assertNoErrors(t)
//	resetTestState()
//}
