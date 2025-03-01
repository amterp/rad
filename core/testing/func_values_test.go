package testing

import "testing"

func TestValues(t *testing.T) {
	rsl := `a = { "alice": "foo", "bob": "bar" }
b = values(a)
print(b)
print(upper(b[0]))
print(values({}))
`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `[ "foo", "bar" ]
FOO
[ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestValues_ErrorsIfGivenString(t *testing.T) {
	rsl := `values("foo")`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L1:8

  values("foo")
         ^^^^^ Got "string" as the 1st argument of values(), but must be: map
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestValues_ErrorsIfGivenNoArgs(t *testing.T) {
	rsl := `values()`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L1:1

  values()
  ^^^^^^^^ values() requires at least 1 argument, but got 0
`
	assertError(t, 1, expected)
	resetTestState()
}

func TestValues_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	rsl := `values({}, {})`
	setupAndRunCode(t, rsl, "--COLOR=never")
	expected := `Error at L1:1

  values({}, {})
  ^^^^^^^^^^^^^^ values() requires at most 1 argument, but got 2
`
	assertError(t, 1, expected)
	resetTestState()
}
