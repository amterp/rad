package testing

import "testing"

func TestKeys(t *testing.T) {
	rsl := `a = { "alice": "foo", "bob": "bar" }
b = keys(a)
print(b)
print(upper(b[0]))
print(keys({}))
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `[ "alice", "bob" ]
ALICE
[ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestKeys_ErrorsIfGivenString(t *testing.T) {
	rsl := `keys("foo")`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L1:6

  keys("foo")
       ^^^^^ Got "string" as the 1st argument of keys(), but must be: map
`
	assertError(t, 1, expected)
}

func TestKeys_ErrorsIfGivenNoArgs(t *testing.T) {
	rsl := `keys()`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L1:1

  keys()
  ^^^^^^ keys() requires at least 1 argument, but got 0
`
	assertError(t, 1, expected)
}

func TestKeys_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	rsl := `keys({}, {})`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L1:1

  keys({}, {})
  ^^^^^^^^^^^^ keys() requires at most 1 argument, but got 2
`
	assertError(t, 1, expected)
}
