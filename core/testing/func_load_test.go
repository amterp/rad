package testing

import "testing"

func Test_Func_Load_Default(t *testing.T) {
	rsl := `
m = {}
x = m.load("k", fn() "first")
x.print()
m.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `first
{ "k": "first" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Load_Cache(t *testing.T) {
	rsl := `
m = {}
m.load("k", fn() "first")
x = m.load("k", fn() "second")
x.print()
m.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `first
{ "k": "first" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Load_Reload(t *testing.T) {
	rsl := `
m = {}
m.load("k", fn() "first")
m.print()
x = m.load("k", fn() "second", reload=true)
x.print()
m.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `{ "k": "first" }
second
{ "k": "second" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Load_Override(t *testing.T) {
	rsl := `
m = {}
m.load("k", fn() "first")
x = m.load("k", fn() exit(1), override="overrode!")
x.print()
m.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `overrode!
{ "k": "overrode!" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Load_DoesNotErrorIfFalseReloadButOverride(t *testing.T) {
	rsl := `
m = {}
m.load("k", fn() "first")
x = m.load("k", fn() exit(1), reload=false, override="overrode!")
x.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `overrode!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Func_Load_ErrorsIfReloadTrueAndOverride(t *testing.T) {
	rsl := `
m = {}
m.load("k", fn() "first")
m.load("k", fn() exit(1), reload=true, override="overrode!")
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L4:3

  m.load("k", fn() exit(1), reload=true, override="overrode!")
    ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    Cannot provide values for both "reload" and "override"
`
	assertError(t, 1, expected)
}
