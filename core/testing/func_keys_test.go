package testing

import "testing"

func TestKeys(t *testing.T) {
	rsl := `a = { "alice": "foo", "bob": "bar" }
print(keys(a))
print(upper(keys(a)[0]))
print(keys({}))
`
	setupAndRunCode(t, rsl)
	expected := `[alice, bob]
ALICE
[]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestKeys_ErrorsIfGivenString(t *testing.T) {
	rsl := `keys("foo")`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L1/4 on 'keys': keys() takes a map, got string\n")
	resetTestState()
}

func TestKeys_ErrorsIfGivenNoArgs(t *testing.T) {
	rsl := `keys()`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L1/4 on 'keys': keys() takes exactly one argument\n")
	resetTestState()
}

func TestKeys_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	rsl := `keys({}, {})`
	setupAndRunCode(t, rsl)
	assertError(t, 1, "RslError at L1/4 on 'keys': keys() takes exactly one argument\n")
	resetTestState()
}
