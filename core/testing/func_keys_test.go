package testing

import "testing"

func TestKeys(t *testing.T) {
	script := `a = { "alice": "foo", "bob": "bar" }
b = keys(a)
print(b)
print(upper(b[0]))
print(keys({}))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "alice", "bob" ]
ALICE
[ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestKeys_ErrorsIfGivenString(t *testing.T) {
	script := `keys("foo")`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30001", "not compatible with expected type 'map'")
}

func TestKeys_ErrorsIfGivenNoArgs(t *testing.T) {
	script := `keys()`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30007", "Missing required argument '_map'")
}

func TestKeys_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	script := `keys({}, {})`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD30007", "Expected at most 1 args, but was invoked with 2")
}
