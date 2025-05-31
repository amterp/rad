package testing

import "testing"

func TestValues(t *testing.T) {
	script := `a = { "alice": "foo", "bob": "bar" }
b = values(a)
print(b)
print(upper(b[0]))
print(values({}))
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "foo", "bar" ]
FOO
[ ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestValues_ErrorsIfGivenString(t *testing.T) {
	script := `values("foo")`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:8

  values("foo")
         ^^^^^ Got "string" as the 1st argument of values(), but must be: map
`
	assertError(t, 1, expected)
}

func TestValues_ErrorsIfGivenNoArgs(t *testing.T) {
	script := `values()`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  values()
  ^^^^^^^^ values() requires at least 1 argument, but got 0
`
	assertError(t, 1, expected)
}

func TestValues_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	script := `values({}, {})`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  values({}, {})
  ^^^^^^^^^^^^^^ values() requires at most 1 argument, but got 2
`
	assertError(t, 1, expected)
}
