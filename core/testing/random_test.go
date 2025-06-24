package testing

import "testing"

func Test_Random_Rand(t *testing.T) {
	script := `seed_random(1)
print(rand())
print(rand())
print(rand())
print(rand())
`
	expected := `0.6046602879796196
0.9405090880450124
0.6645600532184904
0.4377141871869802
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Random_RandInt(t *testing.T) {
	script := `seed_random(1)
print(rand_int(100))
print(rand_int(100))
print(rand_int(100))
print(rand_int(100))
`
	expected := `10
51
21
51
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Random_RandIntMin(t *testing.T) {
	script := `seed_random(1)
print(rand_int(96, 100))
print(rand_int(96, 100))
print(rand_int(96, 100))
print(rand_int(96, 100))
`
	expected := `98
99
97
99
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Random_RandIntNegNumbers(t *testing.T) {
	script := `seed_random(1)
print(rand_int(-10, 10))
print(rand_int(-10, 10))
print(rand_int(-10, 10))
print(rand_int(-10, 10))
`
	expected := `0
1
-9
1
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Random_RandErrorsIfArgs(t *testing.T) {
	script := `rand(1)`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  rand(1)
  ^^^^^^^ Expected at most 0 args, but was invoked with 1
`
	assertError(t, 1, expected)
}

func Test_Random_RandIntErrorsIfNoArgs(t *testing.T) {
	script := `seed_random(1)
rand_int().print()`
	setupAndRunCode(t, script, "--color=never")
	expected := `42983569834913930
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func Test_Random_SeedRandomErrorsIfNoArgs(t *testing.T) {
	script := `seed_random()`
	setupAndRunCode(t, script, "--color=never")
	expected := `Error at L1:1

  seed_random()
  ^^^^^^^^^^^^^ Missing required argument '_seed'
`
	assertError(t, 1, expected)
}

func Test_Random_ErrorsIfMinMaxSame(t *testing.T) {
	script := `rand_int(2, 2)
`
	expected := `Error at L1:1

  rand_int(2, 2)
  ^^^^^^^^^^^^^^ rand_int() min (2) must be less than max (2).
`
	setupAndRunCode(t, script, "--color=never")
	assertError(t, 1, expected)
}
