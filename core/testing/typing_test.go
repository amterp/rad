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

func Test_Typing_StrParamValid(t *testing.T) {
	script := `
foo("hi")
fn foo(x: str):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|hi|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_StrParamInvalid(t *testing.T) {
	script := `
foo(2)
fn foo(x: str):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(2)
      ^ Value '2' (int) is not compatible with expected type 'str'
`
	assertError(t, 1, expected)
}

func Test_Typing_StrReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> str:
	return "hi"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `hi
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_StrReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> str:
	return 2
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '2' (int) is not compatible with expected type 'str'
`
	assertError(t, 1, expected)
}

func Test_Typing_IntParamValid(t *testing.T) {
	script := `
foo(2)
fn foo(x: int):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|2|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_IntParamInvalid(t *testing.T) {
	script := `
foo("2")
fn foo(x: int):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo("2")
      ^^^ Value '"2"' (str) is not compatible with expected type 'int'
`
	assertError(t, 1, expected)
}

func Test_Typing_IntParamNull(t *testing.T) {
	script := `
foo(null)
fn foo(x: int):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(null)
      ^^^^ Value 'null' (null) is not compatible with expected type 'int'
`
	assertError(t, 1, expected)
}

func Test_Typing_IntReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> int:
	return 2
`
	setupAndRunCode(t, script, "--color=never")
	expected := `2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_IntReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> int:
	return "2"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '"2"' (str) is not compatible with expected type 'int'
`
	assertError(t, 1, expected)
}

func Test_Typing_BoolParamValid(t *testing.T) {
	script := `
foo(true)
fn foo(x: bool):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|true|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo should 0 and 1 be allow for bools?
func Test_Typing_BoolParamInvalid(t *testing.T) {
	script := `
foo(1)
fn foo(x: bool):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(1)
      ^ Value '1' (int) is not compatible with expected type 'bool'
`
	assertError(t, 1, expected)
}

func Test_Typing_BoolReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> bool:
	return true
`
	setupAndRunCode(t, script, "--color=never")
	expected := `true
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

// todo should 0 and 1 be allow for bools?
func Test_Typing_BoolReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> bool:
	return 1
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '1' (int) is not compatible with expected type 'bool'
`
	assertError(t, 1, expected)
}

func Test_Typing_ErrorParamValid(t *testing.T) {
	script := `
foo(error("bad!"))
fn foo(x: error):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|bad!|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_ErrorParamInvalid(t *testing.T) {
	script := `
foo(1)
fn foo(x: error):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(1)
      ^ Value '1' (int) is not compatible with expected type 'error'
`
	assertError(t, 1, expected)
}

func Test_Typing_ErrorReturnValid(t *testing.T) {
	script := `
(catch foo()).print()
fn foo() -> error:
	return error("bad!")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `bad!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_ErrorReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> error:
	return 1
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '1' (int) is not compatible with expected type 'error'
`
	assertError(t, 1, expected)
}

func Test_Typing_AnyParamValid(t *testing.T) {
	script := `
foo(1)
foo('hi')
foo(true)
foo(null)
fn foo(x: any):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|1|
|hi|
|true|
|null|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_AnyReturnValid(t *testing.T) {
	script := `
foo(1).print()
foo(2).print()
foo(3).print()
foo(4).print()
fn foo(x) -> any:
    switch x:
        case 1:
			return 1
        case 2:
			return "hi"
        case 3:
			return true
        case 4:
			return null
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1
hi
true
null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_ParamCannotBeVoid(t *testing.T) {
	script := `
foo(1).print()
fn foo(x: void) -> str:
	return "|{x}|"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:8

  fn foo(x: void) -> str:
         ^^ Invalid syntax
`
	assertError(t, 1, expected)
}

func Test_Typing_AnyListParamValid(t *testing.T) {
	script := `
foo([1, 2])
fn foo(x: list):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|[ 1, 2 ]|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_AnyListParamInvalid(t *testing.T) {
	script := `
foo(1)
fn foo(x: list):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(1)
      ^ Value '1' (int) is not compatible with expected type 'list'
`
	assertError(t, 1, expected)
}

func Test_Typing_AnyListParamInvalidVarArg(t *testing.T) {
	script := `
foo(1, 2)
fn foo(x: list):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(1, 2)
      ^ Value '1' (int) is not compatible with expected type 'list'
`
	assertError(t, 1, expected)
}

func Test_Typing_AnyListReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> list:
	return [1, 2]
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_AnyListReturnVarArgReturn(t *testing.T) {
	script := `
foo().print()
fn foo() -> list:
	return 1, 2
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_AnyListReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> list:
	return 1
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '1' (int) is not compatible with expected type 'list'
`
	assertError(t, 1, expected)
}

func Test_Typing_TupleParamValid(t *testing.T) {
	script := `
foo([1, "hi"])
fn foo(x: [int, str]):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|[ 1, "hi" ]|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_TupleParamInvalid(t *testing.T) {
	script := `
foo(["hi", 1])
fn foo(x: [int, str]):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(["hi", 1])
      ^^^^^^^^^
      Value '[ "hi", 1 ]' (list) is not compatible with expected type '[int, str]'
`
	assertError(t, 1, expected)
}

func Test_Typing_TupleReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> [int, str]:
	return [1, "2"]
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, "2" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_TupleReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> [int, str]:
	return [1, 2]
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '[ 1, 2 ]' (list) is not compatible with expected type '[int, str]'
`
	assertError(t, 1, expected)
}

func Test_Typing_ListParamValid(t *testing.T) {
	script := `
foo([1])
foo([1, 2])
fn foo(x: int[]):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|[ 1 ]|
|[ 1, 2 ]|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_ListParamInvalid(t *testing.T) {
	script := `
foo([1, "hi"])
fn foo(x: int[]):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo([1, "hi"])
      ^^^^^^^^^
      Value '[ 1, "hi" ]' (list) is not compatible with expected type 'int[]'
`
	assertError(t, 1, expected)
}

func Test_Typing_ListReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> int[]:
	return [1, 2]
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ 1, 2 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_ListReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> int[]:
	return [1, "2"]
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '[ 1, "2" ]' (list) is not compatible with expected type 'int[]'
`
	assertError(t, 1, expected)
}

func Test_Typing_StrEnumParamValid(t *testing.T) {
	script := `
foo("foo")
fn foo(x: ["foo", "bar"]):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|foo|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_StrEnumParamInvalid(t *testing.T) {
	script := `
foo("quz")
fn foo(x: ["foo", "bar"]):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo("quz")
      ^^^^^ Value '"quz"' (str) is not compatible with expected type 'str enum'
`
	assertError(t, 1, expected)
}

func Test_Typing_StrEnumReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> ["foo", "bar"]:
	return "foo"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `foo
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_StrEnumReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> ["foo", "bar"]:
	return "quz"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '"quz"' (str) is not compatible with expected type 'str enum'
`
	assertError(t, 1, expected)
}

func Test_Typing_AnyMapParamValid(t *testing.T) {
	script := `
foo({"a": 1})
fn foo(x: map):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|{ "a": 1 }|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_AnyMapParamInvalid(t *testing.T) {
	script := `
foo(2)
fn foo(x: map):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(2)
      ^ Value '2' (int) is not compatible with expected type 'map'
`
	assertError(t, 1, expected)
}

func Test_Typing_AnyMapReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> map:
	return {"a": 1}
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "a": 1 }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_AnyMapReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> map:
	return 1
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '1' (int) is not compatible with expected type 'map'
`
	assertError(t, 1, expected)
}

func Test_Typing_StructParamValid(t *testing.T) {
	script := `
foo({"key1": 10})
fn foo(x: {"key1": int}):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|{ "key1": 10 }|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_StructParamInvalid(t *testing.T) {
	script := `
foo({"key2": 10})
fn foo(x: {"key1": int}):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo({"key2": 10})
      ^^^^^^^^^^^^
      Value '{ "key2": 10 }' (map) is not compatible with expected type 'struct'
`
	assertError(t, 1, expected)
}

func Test_Typing_StructReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> {"key1": int}:
	return {"key1": 10}
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "key1": 10 }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_StructReturnInvalidKeys(t *testing.T) {
	script := `
foo().print()
fn foo() -> {"key1": int}:
	return {"key2": 10}
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '{ "key2": 10 }' (map) is not compatible with expected type 'struct'
`
	assertError(t, 1, expected)
}

func Test_Typing_StructReturnInvalidValue(t *testing.T) {
	script := `
foo().print()
fn foo() -> {"key1": int}:
	return {"key2": "hi"}
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^
  Value '{ "key2": "hi" }' (map) is not compatible with expected type 'struct'
`
	assertError(t, 1, expected)
}

func Test_Typing_MapParamValid(t *testing.T) {
	script := `
foo({"key1": 10})
fn foo(x: {str: int}):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|{ "key1": 10 }|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_MapParamInvalid(t *testing.T) {
	script := `
foo({"key1": "hi"})
fn foo(x: {str: int}):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo({"key1": "hi"})
      ^^^^^^^^^^^^^^
      Value '{ "key1": "hi" }' (map) is not compatible with expected type '{ str: int }'
`
	assertError(t, 1, expected)
}

func Test_Typing_MapReturnValid(t *testing.T) {
	script := `
foo().print()
fn foo() -> {str: int}:
	return {"key1": 10}
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ "key1": 10 }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_MapReturnInvalidKeys(t *testing.T) {
	script := `
foo().print()
fn foo() -> {str: int}:
	return {"key1": "hi"}
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^
  Value '{ "key1": "hi" }' (map) is not compatible with expected type '{ str: int }'
`
	assertError(t, 1, expected)
}

func Test_Typing_VarArgParamValid(t *testing.T) {
	script := `
foo()
foo(1)
foo(1, 2)
fn foo(*x: int):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|[ ]|
|[ 1 ]|
|[ 1, 2 ]|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_VarArgParamInvalid(t *testing.T) {
	script := `
foo(1, "hi")
fn foo(*x: int):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:8

  foo(1, "hi")
         ^^^^ Value '"hi"' (str) is not compatible with expected type 'int'
`
	assertError(t, 1, expected)
}

func Test_Typing_VariadicParamNoTyping(t *testing.T) {
	script := `
foo()
foo(1)
foo(1, 2)
fn foo(*x):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|[ ]|
|[ 1 ]|
|[ 1, 2 ]|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_PositionalThenVariadic(t *testing.T) {
	script := `
foo("hi", true)
foo("hi", true, 1)
foo("hi", true, 1, 2)
fn foo(x: str, y: bool, *z: int):
	print("|{x}|{y}|{z}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|hi|true|[ ]|
|hi|true|[ 1 ]|
|hi|true|[ 1, 2 ]|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_PositionalThenVariadicThenNamedOnly(t *testing.T) {
	script := `
foo("hi", true)
foo("hi", true, 1, 2, a=3)
fn foo(x: str, y: bool, *z: int, *, a: int = 0):
	print("|{x}|{y}|{z}|{a}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|hi|true|[ ]|0|
|hi|true|[ 1, 2 ]|3|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_OptionalParamValid(t *testing.T) {
	script := `
foo()
foo(null)
foo(1)
fn foo(x: int?):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|null|
|null|
|1|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_OptionalParamInvalid(t *testing.T) {
	script := `
foo("hi")
fn foo(x: int?):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo("hi")
      ^^^^ Value '"hi"' (str) is not compatible with expected type 'int?'
`
	assertError(t, 1, expected)
}

func Test_Typing_OptionalReturnValid(t *testing.T) {
	script := `
foo().print()
bar().print()
fn foo() -> int?:
	return 1
fn bar() -> int?:
	return null
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1
null
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_OptionalReturnInvalid(t *testing.T) {
	script := `
foo().print()
fn foo() -> int?:
	return "hi"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo().print()
  ^^^^^ Value '"hi"' (str) is not compatible with expected type 'int?'
`
	assertError(t, 1, expected)
}

func Test_Typing_UnionParamValid(t *testing.T) {
	script := `
foo(1)
foo("hi")
fn foo(x: int|str):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `|1|
|hi|
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_UnionParamInvalid(t *testing.T) {
	script := `
foo(1.2)
fn foo(x: int|str):
	print("|{x}|")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  foo(1.2)
      ^^^ Value '1.2' (float) is not compatible with expected type 'int|str'
`
	assertError(t, 1, expected)
}

func Test_Typing_UnionReturnValid(t *testing.T) {
	script := `
foo(1).print()
foo(2).print()
fn foo(x) -> int|str:
	switch x:
		case 1:
			return 1
		case 2:
			return "hi"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `1
hi
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Typing_UnionReturnInvalid(t *testing.T) {
	script := `
foo(1).print()
foo(2).print()
fn foo(x) -> int|str:
	return 1.2
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo(1).print()
  ^^^^^^ Value '1.2' (float) is not compatible with expected type 'int|str'
`
	assertError(t, 1, expected)
}
