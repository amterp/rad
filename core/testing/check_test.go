package testing

import "testing"

func Test_Check_Valid(t *testing.T) {
	// todo should be more happy about it!
	expected := `No diagnostics to report.
`
	setupAndRunArgs(t, "check", "./rad_scripts/hello.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Check(t *testing.T) {
	expected := `L1:9: ERROR

     1 | hello = 2 a
       |         ^ Invalid syntax
       |         (code: RAD10001)

L3:2: ERROR

     3 | 	yes no
       |  ^ Invalid syntax
       |  (code: RAD10001)

Reported 2 diagnostics.
`
	setupAndRunArgs(t, "check", "./rad_scripts/invalid.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertExitCode(t, 1)
}

func Test_Check_FunctionShadowsArgument_Simple(t *testing.T) {
	script := `
args:
    open str?

fn open(path: str) -> void:
    print("Opening {path}")
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L5:4

  fn open(path: str) -> void:
     ^^^^ Hoisted function 'open' shadows an argument with the same name
`
	assertError(t, 1, expected)
}

func Test_Check_FunctionShadowsArgument_Multiple(t *testing.T) {
	script := `
args:
    name str
    count int

fn name() -> str:
    return "test"

fn count() -> int:
    return 10
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L6:4

  fn name() -> str:
     ^^^^ Hoisted function 'name' shadows an argument with the same name
`
	assertError(t, 1, expected)
}

func Test_Check_FunctionShadowsArgument_NoArgsBlock(t *testing.T) {
	script := `
fn open(path: str) -> void:
    print("Opening {path}")

open("test.txt")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Opening test.txt\n")
	assertNoErrors(t)
}

func Test_Check_FunctionShadowsArgument_DifferentNames(t *testing.T) {
	script := `
args:
    name str = "World"

fn greet(person: str) -> void:
    print("Hi {person}!")

greet("Alice")
print("Hello {name}!")
`
	setupAndRunCode(t, script, "Bob", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "Hi Alice!\nHello Bob!\n")
	assertNoErrors(t)
}

func Test_Check_FunctionShadowsArgument_NestedFunctionAllowed(t *testing.T) {
	script := `
args:
    name str = "World"

fn outer() -> void:
    fn name() -> str:
        return "inner"
    print(name())

outer()
print("Hello {name}!")
`
	setupAndRunCode(t, script, "Bob", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "inner\nHello Bob!\n")
	assertNoErrors(t)
}
