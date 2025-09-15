package testing

import "testing"

func Test_Fn_ReturnInForLoop(t *testing.T) {
	script := `
fn foo():
    for i in range(5):
        return 5

a = foo()
print(":{a}:")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ":5:\n")
	assertNoErrors(t)
}

func Test_Fn_ReturnInWhileLoop(t *testing.T) {
	script := `
fn bar():
    i = 0
    while i < 5:
        return 10
        i = i + 1

b = bar()
print(":{b}:")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ":10:\n")
	assertNoErrors(t)
}

func Test_Fn_ReturnDirectly_ShouldWork(t *testing.T) {
	script := `
fn direct():
    return 42

c = direct()
print(":{c}:")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ":42:\n")
	assertNoErrors(t)
}

func Test_Fn_YieldInForLoopInSwitch(t *testing.T) {
	script := `
result = switch "test":
    case "test":
        for i in range(5):
            if i == 2:
                yield 42
    default:
        yield -1

print(":{result}:")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ":42:\n")
	assertNoErrors(t)
}

func Test_Fn_YieldInNestedLoops(t *testing.T) {
	script := `
result = switch "nested":
    case "nested":
        for i in range(3):
            for j in range(3):
                if i == 1 and j == 1:
                    yield "success"
    default:
        yield "fail"

print(":{result}:")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, ":success:\n")
	assertNoErrors(t)
}
