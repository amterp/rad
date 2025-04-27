package testing

import "testing"

func Test_Fn_SingleLiners(t *testing.T) {
	rsl := `
foo = fn() 5
bar = fn(x) x * 2
quz = fn(x, y) x * y

foo().print()
bar(5).print()
quz(4, 10).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n10\n40\n")
	assertNoErrors(t)
}

func Test_Fn_CanCopyBuiltin(t *testing.T) {
	t.Skip("TODO not yet expected to pass")
	rsl := `
foo = upper
"test".foo().print()
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "[ 3, 2 ]\n")
	assertNoErrors(t)
}

func Test_Fn_SingleLineClosure(t *testing.T) {
	rsl := `
foo = fn(b) a * b
a = 2
foo(10).print()
a = 5
foo(10).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "20\n50\n")
	assertNoErrors(t)
}

func Test_Fn_Block(t *testing.T) {
	rsl := `
foo = fn():
	return 5
bar = fn(x):
    out = x * 2
	return out
quz = fn(x, y):
	return x * y

foo().print()
bar(5).print()
quz(4, 10).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n10\n40\n")
	assertNoErrors(t)
}

func Test_Fn_BlockClosure(t *testing.T) {
	rsl := `
a = 2
foo = fn(x):
	return a * x

foo(5).print()

a = 5
foo(5).print()
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10\n25\n")
	assertNoErrors(t)
}

func Test_Fn_BlockClosureWithinFn(t *testing.T) {
	rsl := `
a = 2
foo = fn(x):
	a = 3
	return a * x

a = 4
foo(5).print()
a.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "15\n4\n")
	assertNoErrors(t)
}

func Test_Fn_CanPrint(t *testing.T) {
	rsl := `
foo = fn() 5
foo.print()
bar = fn(x, y) x * y
bar.print()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `<fn ()>
<fn (x, y)>
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
