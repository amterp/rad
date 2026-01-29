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

func Test_Fallback_MapBracketSyntax(t *testing.T) {
	script := `
m = {"a": 1, "b": 2}
print(m["c"] ?? "default")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "default\n")
	assertNoErrors(t)
}

func Test_Fallback_MapDotSyntax(t *testing.T) {
	script := `
m = {"a": 1, "b": 2}
print(m.c ?? "default")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "default\n")
	assertNoErrors(t)
}

func Test_Fallback_MapChained(t *testing.T) {
	script := `
m = {"a": 1}
print(m["b"] ?? m["c"] ?? "default")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "default\n")
	assertNoErrors(t)
}

func Test_Fallback_MapNestedInnerKeyMissing(t *testing.T) {
	script := `
m = {"a": {"x": 1}}
print(m["a"]["y"] ?? "default")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "default\n")
	assertNoErrors(t)
}

func Test_Fallback_MapNestedOuterKeyMissing(t *testing.T) {
	script := `
m = {"a": {"x": 1}}
print(m["b"]["y"] ?? "default")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "default\n")
	assertNoErrors(t)
}

func Test_Fallback_MapKeyExists(t *testing.T) {
	script := `
m = {"a": 1, "b": 2}
print(m["a"] ?? "default")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "1\n")
	assertNoErrors(t)
}

func Test_Fallback_MapFallbackAlsoErrors(t *testing.T) {
	script := `
m = {"a": 1}
n = {"x": 2}
print(m["b"] ?? n["y"] ?? "final")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "final\n")
	assertNoErrors(t)
}

func Test_Fallback_MapNullValueVsMissingKey(t *testing.T) {
	script := `
m = {"exists": null, "other": 1}
print(m["exists"] ?? "fallback")
print(m["missing"] ?? "fallback")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `null
fallback
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
