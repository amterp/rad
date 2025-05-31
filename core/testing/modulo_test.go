package testing

import "testing"

func Test_Modulo_PositiveInts(t *testing.T) {
	script := `
print(4 % 1)
print(4 % 2)
print(4 % 3)
print(4 % 4)
print(4 % 5)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0
0
1
0
4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Modulo_NegativeInts(t *testing.T) {
	script := `
print(-1 % 3)
print(-2 % 3)
print(-3 % 3)
print(-4 % 3)

print(3 % -1)
print(3 % -2)
print(3 % -3)
print(3 % -4)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `-1
-2
0
-1
0
1
0
3
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Modulo_PositiveFloats(t *testing.T) {
	script := `
print(1.25 % 0.5)
print(5.0 % 2.5)
print(5.0 % 2.0)
print(10.0 % 3.0)
print(10.0 % 3.3)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0.25
0
1
1
0.10000000000000053
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Modulo_NegativeFloats(t *testing.T) {
	script := `
print(-1.25 % 0.5)
print(-5.0 % 2.5)
print(-5.0 % 2.0)
print(-10.0 % 3.0)
print(-10.0 % 3.3)

print(1.25 % -0.5)
print(5.0 % -2.5)
print(5.0 % -2.0)
print(10.0 % -3.0)
print(10.0 % -3.3)

print(-1.25 % -0.5)
print(-5.0 % -2.5)
print(-5.0 % -2.0)
print(-10.0 % -3.0)
print(-10.0 % -3.3)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `-0.25
-0
-1
-1
-0.10000000000000053
0.25
0
1
1
0.10000000000000053
-0.25
-0
-1
-1
-0.10000000000000053
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Modulo_CompoundOperatorInts(t *testing.T) {
	script := `
a = 10
a %= 6
print(a)

b = -10
b %= 6
print(b)

c = 10
c %= -6
print(c)

d = -10
d %= -6
print(d)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `4
-4
4
-4
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Modulo_CompoundOperatorFloats(t *testing.T) {
	script := `
a = 7.2
a %= 3.5
print(a)

b = -7.2
b %= 3.5
print(b)

c = 7.2
c %= -3.5
print(c)

d = -7.2
d %= -3.5
print(d)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `0.20000000000000018
-0.20000000000000018
0.20000000000000018
-0.20000000000000018
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Modulo_PositiveIntModulo0Errors(t *testing.T) {
	script := `
print(5 % 0)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:11

  print(5 % 0)
            ^ Value is 0, cannot modulo by 0
`
	assertError(t, 1, expected)
}

func Test_Modulo_NegativeIntModulo0Errors(t *testing.T) {
	script := `
print(-5 % 0)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:12

  print(-5 % 0)
             ^ Value is 0, cannot modulo by 0
`
	assertError(t, 1, expected)
}

func Test_Modulo_PositiveFloatModulo0Errors(t *testing.T) {
	script := `
print(5.5 % 0)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:13

  print(5.5 % 0)
              ^ Value is 0, cannot modulo by 0
`
	assertError(t, 1, expected)
}

func Test_Modulo_NegativeFloatModulo0Errors(t *testing.T) {
	script := `
print(-5.5 % 0)
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L2:14

  print(-5.5 % 0)
               ^ Value is 0, cannot modulo by 0
`
	assertError(t, 1, expected)
}

func Test_Modulo_CompoundModulo0Errors(t *testing.T) {
	script := `
a = 5
a %= 0
`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L3:6

  a %= 0
       ^ Value is 0, cannot modulo by 0
`
	assertError(t, 1, expected)
}
