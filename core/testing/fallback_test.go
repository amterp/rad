package testing

import "testing"

func Test_Fallback_SimpleNoError(t *testing.T) {
	script := `
a = parse_int("2") ?? 10
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fallback_SimpleError(t *testing.T) {
	script := `
a = parse_int("two") ?? 10
print(a)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fallback_CanInvokeFunction(t *testing.T) {
	script := `
a = parse_int("two") ?? foo()
print(a)

fn foo():
	print("ran foo")
	return 10
`
	setupAndRunCode(t, script, "--color=never")
	expected := `ran foo
10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fallback_LazyEval(t *testing.T) {
	script := `
a = parse_int("2") ?? foo()
print(a)

fn foo():
	print("ran foo")
	return 10
`
	setupAndRunCode(t, script, "--color=never")
	expected := `2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fallback_CanMixWithCatch(t *testing.T) {
	script := `
a = parse_int("two") ?? foo() catch:
	print("caught error {a}")

fn foo():
	return error("bad!")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `caught error bad!
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
