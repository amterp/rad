package testing

import "testing"

func Test_Func_Map_ListLambda(t *testing.T) {
	rsl := `
a = ["alice", "bob", "charlie"]
a.map(fn(n) n.upper()).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "ALICE", "BOB", "CHARLIE" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Map_ListFn(t *testing.T) {
	rsl := `
a = ["alice", "bob", "charlie"]
to_upper = fn(n) n.upper()
a.map(to_upper).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "ALICE", "BOB", "CHARLIE" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Map_MapLambda(t *testing.T) {
	rsl := `
a = { "alice": "bobson", "charlie": "davidson" }
a.map(fn(k, v) "{k} {v}").print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "alice bobson", "charlie davidson" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Map_MapFn(t *testing.T) {
	rsl := `
a = { "alice": "bobson", "charlie": "davidson" }
joiner = fn(k, v) "{k} {v}"
a.map(joiner).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "alice bobson", "charlie davidson" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo additional functions
//  - mapKeys, mapValues
//  - mapToKeys, mapToValues (take both k, v)
//  - though, first check alternatives & modern solutions
