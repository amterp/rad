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
         ^^^^^ Value '"foo"' (str) is not compatible with expected type 'map'
`
	assertError(t, 1, expected)
}

func TestValues_ErrorsIfGivenNoArgs(t *testing.T) {
	script := `values()`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  values()
  ^^^^^^^^ Missing required argument '_map'
`
	assertError(t, 1, expected)
}

func TestValues_ErrorsIfGivenMoreThanOneArg(t *testing.T) {
	script := `values({}, {})`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  values({}, {})
  ^^^^^^^^^^^^^^ Expected at most 1 args, but was invoked with 2
`
	assertError(t, 1, expected)
}
