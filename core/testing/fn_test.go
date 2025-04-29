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

func Test_Fn_CanCopyBuiltIn(t *testing.T) {
	rsl := `
foo = upper
"test".foo().print()
`
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "TEST\n")
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

func Test_Fn_CanMultiReturn(t *testing.T) {
	rsl := `
foo = fn() (1, 2)
a, b = foo()
print(a, b)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `1 2
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_Increment(t *testing.T) {
	rsl := `
c = 0
foo = fn() c++
for _ in range(10):
	foo()
print(c)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `10
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_CompoundIncr(t *testing.T) {
	rsl := `
c = 0
foo = fn() c += 2
for _ in range(10):
	foo()
print(c)
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `20
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_RadBlockLambda(t *testing.T) {
	rsl := `
a = { "name": "alex" }

name = json.name
display a:
    fields name
    name:
        map fn(n) n.upper()
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `name 
ALEX  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_RadBlockIdentifier(t *testing.T) {
	rsl := `
a = { "name": "alex" }

name = json.name
display a:
    fields name
    name:
        map upper
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `name 
ALEX  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Fn_RadBlockBlock(t *testing.T) {
	rsl := `
a = { "name": "alex" }

name = json.name
display a:
    fields name
    name:
        map fn(n):
            out = n.upper()
            return out
`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `name 
ALEX  
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
