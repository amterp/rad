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
	setupAndRunCode(t, script)
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
	setupAndRunCode(t, script)
	expected := `Error at L2:5

  a = foo()
      ^^^^^ this is an error
`
	assertError(t, 1, expected)
}

func Test_Catch_CanCatchOnNestedFunctions(t *testing.T) {
	script := `
a = catch foo()
print("First", a)
a = catch foo().foo()
print("Second", a)

fn foo(x):
	if type_of(x) == "error":
		return error("this is an error: {x}")
	return 1
`
	setupAndRunCode(t, script)
	expected := `
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ImmediatelyPropagatesOnNestedFunctions(t *testing.T) {
	script := `
foo().foo()

fn foo(x):
	if type_of(x) == "error":
		return error("two")
	return error("one")
`
	setupAndRunCode(t, script)
	expected := `one, not two
`
	assertError(t, 1, expected)
}

func Test_Catch_CanCatchOnFromTernary(t *testing.T) {
	script := `
a = catch true ? foo(1) : foo(2)
print("Got: {a}")

fn foo(x):
	print("Running {x}"
	return error("error: {x}")
`
	setupAndRunCode(t, script)
	expected := `Running 1
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
	setupAndRunCode(t, script)
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
	setupAndRunCode(t, script)
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
	setupAndRunCode(t, script)
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
	setupAndRunCode(t, script)
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
	setupAndRunCode(t, script)
	expected := `bar error
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
	setupAndRunCode(t, script)
	expected := `foo error
`
	assertError(t, 1, expected)
}

func Test_Catch_LambdaMapErrors(t *testing.T) {
	script := `
a = [1, 2]
a.map(foo).print()

foo = fn(x):
	return error("this is an error")
`
	setupAndRunCode(t, script)
	expected := `
`
	assertError(t, 1, expected)
}

func Test_Catch_LambdaMapCanCatch(t *testing.T) {
	script := `
a = [1, 2]
a.map(foo).print()

foo = fn(x):
	return error("this is an error")
`
	setupAndRunCode(t, script)
	expected := `
`
	assertError(t, 1, expected)
}
