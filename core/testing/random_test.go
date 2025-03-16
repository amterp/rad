package testing

import "testing"

func TestRandom_Rand(t *testing.T) {
	rsl := `seed_random(1)
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
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRandom_RandInt(t *testing.T) {
	rsl := `seed_random(1)
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
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRandom_RandIntMin(t *testing.T) {
	rsl := `seed_random(1)
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
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRandom_RandIntNegNumbers(t *testing.T) {
	rsl := `seed_random(1)
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
	setupAndRunCode(t, rsl, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}

func TestRandom_RandErrorsIfArgs(t *testing.T) {
	rsl := `rand(1)`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L1:1

  rand(1)
  ^^^^^^^ rand() requires at most 0 arguments, but got 1
`
	assertError(t, 1, expected)
}

func TestRandom_RandIntErrorsIfNoArgs(t *testing.T) {
	rsl := `rand_int()`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L1:1

  rand_int()
  ^^^^^^^^^^ rand_int() requires at least 1 argument, but got 0
`
	assertError(t, 1, expected)
}

func TestRandom_SeedRandomErrorsIfNoArgs(t *testing.T) {
	rsl := `seed_random()`
	setupAndRunCode(t, rsl, "--color=never")
	expected := `Error at L1:1

  seed_random()
  ^^^^^^^^^^^^^ seed_random() requires at least 1 argument, but got 0
`
	assertError(t, 1, expected)
}
