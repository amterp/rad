package testing

import (
	"testing"
)

func Test_Catch_CanCatch(t *testing.T) {
	script := `
a = catch foo()
print("Got: {a}")

fn foo():
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Got: this is an error
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ErrorsIfNoCatch(t *testing.T) {
	script := `
a = foo()
print("Got: {a}")

fn foo():
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:5

  a = foo()
      ^^^^^ this is an error
`
	assertError(t, 1, expected)
}

func Test_Catch_CanCatchOnNestedFunctions(t *testing.T) {
	script := `
a = catch foo(1)
print("First", a)
a = catch foo(2).foo()
print("Second", a)

fn foo(x):
	print("Foo!", x)
	out = x
	if x == 2:
		out = error("this is an error: {x}")
	return out
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Foo! 1
First 1
Foo! 2
Second this is an error: 2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ImmediatelyPropagatesOnNestedFunctions(t *testing.T) {
	script := `
foo(1).foo()

fn foo(x):
  print_err("Foo!", x)
  return error(out)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Foo! 1
Error at L6:16

    return error(out)
                 ^^^ Undefined variable: out
`
	assertError(t, 1, expected)
}

func Test_Catch_CanCatchOnFromTernary(t *testing.T) {
	script := `
a = catch true ? foo(1) : foo(2)
print("Got: {a}")

fn foo(x):
	print("Running {x}")
	return error("error: {x}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Running 1
Got: error: 1
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ListComprehensionPropagates(t *testing.T) {
	script := `
a = bar()

fn foo(x):
	print_err("Foo {x}")
	return error("error {x}")

fn bar():
	print_err("bar")
	return [foo(a) for a in [1, 2]]
`
	setupAndRunCode(t, script, "--color=never")
	expected := `bar
Foo 1
Error at L10:10

  	return [foo(a) for a in [1, 2]]
           ^^^^^^ error 1
`
	assertOnlyOutput(t, stdErrBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Catch_ListComprehensionCanCatch(t *testing.T) {
	script := `
a = catch [foo(a) for a in [1, 2]]
print("Got: {a}")

fn foo(x):
	print("Foo {x}")
	return error("error: {x}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Foo 1
Got: error: 1
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ErrorsInList(t *testing.T) {
	script := `
a = [foo()]

fn foo():
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:6

  a = [foo()]
       ^^^^^ this is an error
`
	assertError(t, 1, expected)
}

func Test_Catch_CanCatchInList(t *testing.T) {
	script := `
a = [catch foo()]
print(a)

fn foo():
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `[ "this is an error" ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ErrorsInMap(t *testing.T) {
	script := `
a = {1: foo()}

fn foo():
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:9

  a = {1: foo()}
          ^^^^^ this is an error
`
	assertError(t, 1, expected)
}

func Test_Catch_CanCatchInMap(t *testing.T) {
	script := `
a = {1: catch foo()}
print(a)

fn foo():
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `{ 1: "this is an error" }
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_CanPropagate(t *testing.T) {
	script := `
foo()

fn foo():
	a = bar()
	return error("foo error")

fn bar():
	return error("bar error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L5:6

  	a = bar()
       ^^^^^ bar error
`
	assertError(t, 1, expected)
}

func Test_Catch_CanCatchInFunction(t *testing.T) {
	script := `
foo()

fn foo():
	a = catch bar()
	return error("foo error")

fn bar():
	return error("bar error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:1

  foo()
  ^^^^^ foo error
`
	assertError(t, 1, expected)
}

func Test_Catch_LambdaMapErrors(t *testing.T) {
	script := `
a = [1, 2]
a.map(foo).print()

fn foo(x):
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:3

  a.map(foo).print()
    ^^^^^^^^ this is an error
`
	assertError(t, 1, expected)
}

func Test_Catch_LambdaMapCanCatch(t *testing.T) {
	t.Skipf("TODO: not possible yet. Probably just want a catch=true named arg in map?")
	script := `
a = [1, 2]
a.map(foo).print()

foo = fn(x):
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `
`
	assertError(t, 1, expected)
}
