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
	expected := `Error at L1:6

  keys("foo")
       ^^^^^ Value '"foo"' (str) is not compatible with expected type 'map'
`
	assertError(t, 1, expected)
}

func TestKeys_ErrorsIfGivenNoArgs(t *testing.T) {
	script := `keys()`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  keys()
  ^^^^^^ Missing required argument '_map'
`
	assertError(t, 1, expected)
}

func TestKeys_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	script := `keys({}, {})`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  keys({}, {})
  ^^^^^^^^^^^^ Expected at most 1 args, but was invoked with 2
`
	assertError(t, 1, expected)
}
