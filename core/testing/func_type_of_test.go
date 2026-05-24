package testing

import "testing"

// Pins the full enum returned by `type_of` so static narrowing (Phase 4
// onward) can rely on the values without silent gaps.

func Test_TypeOf_Int(t *testing.T) {
	setupAndRunCode(t, `print(type_of(42))`, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "int\n")
	assertNoErrors(t)
}

func Test_TypeOf_Float(t *testing.T) {
	setupAndRunCode(t, `print(type_of(3.14))`, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "float\n")
	assertNoErrors(t)
}

func Test_TypeOf_Str(t *testing.T) {
	setupAndRunCode(t, `print(type_of("hi"))`, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "str\n")
	assertNoErrors(t)
}

func Test_TypeOf_Bool(t *testing.T) {
	setupAndRunCode(t, `print(type_of(true))`, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "bool\n")
	assertNoErrors(t)
}

func Test_TypeOf_List(t *testing.T) {
	setupAndRunCode(t, `print(type_of([1, 2]))`, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "list\n")
	assertNoErrors(t)
}

func Test_TypeOf_Map(t *testing.T) {
	setupAndRunCode(t, `print(type_of({"a": 1}))`, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "map\n")
	assertNoErrors(t)
}

func Test_TypeOf_Null(t *testing.T) {
	setupAndRunCode(t, `print(type_of(null))`, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "null\n")
	assertNoErrors(t)
}

func Test_TypeOf_Error(t *testing.T) {
	// parse_int on a non-numeric string yields an `error` value;
	// `catch` captures it into the local for inspection.
	script := `
e = parse_int("nope") catch:
    pass
print(type_of(e))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "error\n")
	assertNoErrors(t)
}

func Test_TypeOf_Function(t *testing.T) {
	script := `
f = fn() 1
print(type_of(f))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "function\n")
	assertNoErrors(t)
}
