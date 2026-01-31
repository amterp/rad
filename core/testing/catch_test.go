package testing

import (
	"testing"
)

func Test_Catch_CanCatch(t *testing.T) {
	script := `
a = foo() catch:
	pass
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
	assertErrorContains(t, 1, "RAD20000", "this is an error")
}

func Test_Catch_CanCatchOnNestedFunctions(t *testing.T) {
	script := `
a = foo(1) catch:
	pass
print("First", a)
a = foo(2).foo() catch:
	pass
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
	assertErrorContains(t, 1, "RAD20028", "Undefined variable: out")
}

func Test_Catch_CanCatchOnFromTernary(t *testing.T) {
	script := `
a = (true ? foo(1) : foo(2)) catch:
	pass
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
	assertErrorContains(t, 1, "RAD20000", "error 1")
}

func Test_Catch_ListComprehensionCanCatch(t *testing.T) {
	script := `
a = [foo(a) for a in [1, 2]] catch:
	pass
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
	assertErrorContains(t, 1, "RAD20000", "this is an error")
}

func Test_Catch_ErrorsInMap(t *testing.T) {
	script := `
a = {1: foo()}

fn foo():
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "this is an error")
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
	assertErrorContains(t, 1, "RAD20000", "bar error")
}

func Test_Catch_CanCatchInFunction(t *testing.T) {
	script := `
foo()

fn foo():
	a = bar() catch:
		pass
	return error("foo error {a}")

fn bar():
	return error("bar error")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "foo error bar error")
}

func Test_Catch_LambdaMapErrors(t *testing.T) {
	script := `
a = [1, 2]
a.map(foo).print()

fn foo(x):
	return error("this is an error")
`
	setupAndRunCode(t, script, "--color=never")
	assertErrorContains(t, 1, "RAD20000", "this is an error")
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

// Test control flow propagation from catch blocks

func Test_Catch_ReturnInCatchBlock(t *testing.T) {
	script := `
result = foo()
print("Result: {result}")

fn foo():
	for i in range(10):
		a = bar(i) catch:
			return "caught error at {i}"
	return "completed all iterations"

fn bar(x):
	if x == 3:
		return error("error at {x}")
	return x
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Result: caught error at 3
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_BreakInCatchBlock(t *testing.T) {
	script := `
results = []
for i in range(10):
	a = foo(i) catch:
		print("Caught error at {i}, breaking")
		break
	results += [a]

print("Results: {results}")

fn foo(x):
	if x == 5:
		return error("error at {x}")
	return x
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Caught error at 5, breaking
Results: [ 0, 1, 2, 3, 4 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ContinueInCatchBlock(t *testing.T) {
	script := `
results = []
for i in range(5):
	a = foo(i) catch:
		print("Caught error at {i}, continuing")
		continue
	results += [a]

print("Results: {results}")

fn foo(x):
	if x == 2:
		return error("error at {x}")
	return x * 10
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Caught error at 2, continuing
Results: [ 0, 10, 30, 40 ]
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_YieldInCatchBlock(t *testing.T) {
	script := `
result = switch "a":
	case "a":
		parse_int("two") catch:
			yield 5
print(result)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `5
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_ExprStmtReturnsVoidNotError(t *testing.T) {
	script := `
// expr_stmt with catch should return void, not the error value
foo() catch:
	print("Caught error")

print("Continued execution")

fn foo():
	return error("test error")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Caught error
Continued execution
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Catch_AssignWithCatchControlFlow(t *testing.T) {
	script := `
fn test():
	for i in range(5):
		err = foo(i) catch:
			if i == 2:
				return "early return from catch"
			print("Caught: {err}")
	return "completed"

result = test()
print("Final: {result}")

fn foo(x):
	if x == 1 or x == 2:
		return error("error {x}")
	return "ok {x}"
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Caught: error 1
Final: early return from catch
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
