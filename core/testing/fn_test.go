package testing

import "testing"

func Test_Fn_SingleLiners(t *testing.T) {
	script := `
foo = fn() 5
bar = fn(x) x * 2
quz = fn(x, y) x * y

foo().print()
bar(5).print()
quz(4, 10).print()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n10\n40\n")
	assertNoErrors(t)
}

func Test_Fn_CanCopyBuiltIn(t *testing.T) {
	script := `
foo = upper
"test".foo().print()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "TEST\n")
	assertNoErrors(t)
}

func Test_Fn_SingleLineClosure(t *testing.T) {
	script := `
foo = fn(b) a * b
a = 2
foo(10).print()
a = 5
foo(10).print()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "20\n50\n")
	assertNoErrors(t)
}

func Test_Fn_Block(t *testing.T) {
	script := `
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
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "5\n10\n40\n")
	assertNoErrors(t)
}

func Test_Fn_BlockClosure(t *testing.T) {
	script := `
a = 2
foo = fn(x):
	return a * x

foo(5).print()

a = 5
foo(5).print()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "10\n25\n")
	assertNoErrors(t)
}

func Test_Fn_BlockClosureWithinFn(t *testing.T) {
	script := `
a = 2
foo = fn(x):
	a = 3
	return a * x

a = 4
foo(5).print()
a.print()
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "15\n4\n")
	assertNoErrors(t)
}

// TODO 2025-06-18: this test regressed, output should provide more signature information
func Test_Fn_CanPrint(t *testing.T) {
	script := `
foo = fn() 5
foo.print()
bar = fn(x, y) x * y
bar.print()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `<fn>
<fn>
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_Increment(t *testing.T) {
	script := `
c = 0
foo = fn() c++
for _ in range(10):
	foo()
print(c)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_CompoundIncr(t *testing.T) {
	script := `
c = 0
foo = fn() c += 2
for _ in range(10):
	foo()
print(c)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `20
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_RadBlockLambda(t *testing.T) {
	script := `
a = { "name": "alex" }

name = json.name
display a:
    fields name
    name:
        map fn(n) n.upper()
`
	setupAndRunCode(t, script, "--color=never")
	expected := `name 
ALEX  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_RadBlockIdentifier(t *testing.T) {
	script := `
a = { "name": "alex" }

name = json.name
display a:
    fields name
    name:
        map upper
`
	setupAndRunCode(t, script, "--color=never")
	expected := `name 
ALEX  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_RadBlockBlock(t *testing.T) {
	script := `
a = { "name": "alex" }

name = json.name
display a:
    fields name
    name:
        map fn(n):
            out = n.upper()
            return out
`
	setupAndRunCode(t, script, "--color=never")
	expected := `name 
ALEX  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_BlockRequiresReturn(t *testing.T) {
	script := `
print(add(4, 5))
fn add(x, y):
	x + y
`
	setupAndRunCode(t, script, "--color=never")
	expected := `
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
