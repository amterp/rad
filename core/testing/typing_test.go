package testing

import "testing"

func Test_Typing_CorrectPositionalArgsAndReturn(t *testing.T) {
	script := `
add(1, 2).print()
fn add(x: float, y: float) -> float:
	return x + y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `3
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_WrongPositionalArgs(t *testing.T) {
	script := `
add(1, "2").print()
fn add(x: float, y: float) -> float:
	return x + y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:8

  add(1, "2").print()
         ^^^ Value '"2"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}

func Test_Typing_WrongReturn(t *testing.T) {
	script := `
add(1, 2).print()
fn add(x: float, y: float) -> float:
	return "hi"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  add(1, 2).print()
  ^^^^^^^^^ Value '"hi"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}

func Test_Typing_VoidReturn(t *testing.T) {
	script := `
foo(1)
bar(2)
fn foo(x: float):
	print("foo!")
fn bar(x: float) -> void:
	print("bar!")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `foo!
bar!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_NamedArg(t *testing.T) {
	script := `
foo(1, 2).print()
foo(x=1, y=2).print()
foo(y=1, x=2).print()
fn foo(x: float, y: float) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0.5
0.5
2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_NamedWrongType(t *testing.T) {
	script := `
foo(x=1, y="2").print()
fn foo(x: float, y: float) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:12

  foo(x=1, y="2").print()
             ^^^ Value '"2"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}

func Test_Typing_NamedUnknown(t *testing.T) {
	script := `
foo(x=1, y=2, z=3).print()
fn foo(x: float, y: float) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:15

  foo(x=1, y=2, z=3).print()
                ^ Unknown named argument 'z'
`
	assertError(t, 1, expected)
}

func Test_Typing_NamedMissing(t *testing.T) {
	script := `
foo(x=1).print()
fn foo(x: float, y: float) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo(x=1).print()
  ^^^^^^^^ Missing required argument 'y'
`
	assertError(t, 1, expected)
}

func Test_Typing_NamedMissingDefaults(t *testing.T) {
	script := `
foo(x=1).print()
fn foo(x: float, y: float = 2) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0.5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_InvalidDefault(t *testing.T) {
	script := `
foo(x=1).print()
fn foo(x: float, y: float = "2") -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:29

  fn foo(x: float, y: float = "2") -> float:
                              ^^^
                              Value '"2"' (str) is not compatible with expected type 'float'
`
	assertError(t, 1, expected)
}

func Test_Typing_TooManyPositional(t *testing.T) {
	script := `
foo(1, 2, 3).print()
fn foo(x: float, y: float) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo(1, 2, 3).print()
  ^^^^^^^^^^^^ Expected at most 2 args, but was invoked with 3
`
	assertError(t, 1, expected)
}

func Test_Typing_TooManyPositionalButNamedOnlyAllowed(t *testing.T) {
	script := `
foo(1, 2, 3).print()
fn foo(x: float, y: float, *, z: float) -> float:
	return x / y + z
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:11

  foo(1, 2, 3).print()
            ^ Too many positional args, remaining args are named-only.
`
	assertError(t, 1, expected)
}

func Test_Typing_PositionalOnly(t *testing.T) {
	script := `
foo(_x=1, _y=2).print()
fn foo(_x: float, _y: float) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(_x=1, _y=2).print()
      ^^ Argument '_x' cannot be passed as named arg, only positionally.
`
	assertError(t, 1, expected)
}

func Test_Typing_OptionalArgNulls(t *testing.T) {
	script := `
foo(x=1)
fn foo(x: float, y?):
	print(x, y)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1 null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_OptionalNamedArgNulls(t *testing.T) {
	script := `
foo(x=1)
fn foo(x: float, *, y?):
	print(x, y)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1 null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_NonVoidFuncErrorsIfNothingReturns(t *testing.T) {
	script := `
foo(x=1)
fn foo(x: float, y?) -> float:
	a = 1
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo(x=1)
  ^^^^^^^^ Expected 'float', but got void value.
`
	assertError(t, 1, expected)
}

func Test_Typing_RepeatArg(t *testing.T) {
	script := `
foo(1, x=2)
fn foo(x: float, y?) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:8

  foo(1, x=2)
         ^ Argument 'x' already specified.
`
	assertError(t, 1, expected)
}

func Test_Typing_RepeatNamedArg(t *testing.T) {
	script := `
foo(1, y = 2, y = 3)
fn foo(x: float, y?) -> float:
	return x / y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:15

  foo(1, y = 2, y = 3)
                ^ Duplicate named argument: y
`
	assertError(t, 1, expected)
}
