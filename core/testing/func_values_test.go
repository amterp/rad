package testing

import "testing"

func TestValues(t *testing.T) {
	rsl := `a = { "alice": "foo", "bob": "bar" }
print(values(a))
print(upper(values(a)[0]))
print(values({}))
`
	setupAndRunCode(t, rsl)
	expected := `[foo, bar]
FOO
[]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestValues_ErrorsIfGivenString(t *testing.T) {
	rsl := `values("foo")`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L1/6 on 'values': values() takes a map, got string\n")
	resetTestState()
}

func TestValues_ErrorsIfGivenNoArgs(t *testing.T) {
	rsl := `values()`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L1/6 on 'values': values() takes exactly one argument\n")
	resetTestState()
}

func TestValues_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	rsl := `values({}, {})`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L1/6 on 'values': values() takes exactly one argument\n")
	resetTestState()
}
