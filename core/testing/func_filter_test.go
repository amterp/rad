package testing

import "testing"

func Test_Func_Filter_ListLambda(t *testing.T) {
	rsl := `
a = ["alice", "bob", "charlie"]
a.filter(fn(n) n.len() > 4).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "alice", "charlie" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Filter_ListFn(t *testing.T) {
	rsl := `
a = ["alice", "bob", "charlie"]
long = fn(n) n.len() > 4
a.filter(long).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "alice", "charlie" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Filter_MapLambda(t *testing.T) {
	rsl := `
a = { "alice": "bobson", "charlie": "davidson" }
a.filter(fn(k, v) k.len() > 5).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "charlie": "davidson" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Filter_MapFn(t *testing.T) {
	rsl := `
a = { "alice": "bobson", "charlie": "davidson" }
long = fn(k, v) k.len() > 5
a.filter(long).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "charlie": "davidson" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Filter_CanChainWithMap(t *testing.T) {
	rsl := `
a = { "alice": "bobson", "charlie": "davidson" }
a.filter(fn(k, v) k.len() > 5).map(fn(k, v) "{k} {v.upper()}").print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "charlie DAVIDSON" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
