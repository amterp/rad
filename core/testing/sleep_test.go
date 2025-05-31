package testing

import "testing"

func TestSleep_IntBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep(10)`)
	assertSleptMillis(t, 10000)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}

func TestSleep_FloatBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep(10.2)`)
	assertSleptMillis(t, 10200)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}

func TestSleep_IntInStringBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep("10")`)
	assertSleptMillis(t, 10000)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}

func TestSleep_FloatInStringBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep("10.2")`)
	assertSleptMillis(t, 10200)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}

func TestSleep_AllowsZero(t *testing.T) {
	setupAndRunCode(t, `sleep(0)`)
	assertSleptMillis(t, 0)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}

func TestSleep_GoCompatibleHumanString(t *testing.T) {
	script := `
sleep("10s")
sleep("10.2s")
sleep("12345ms")
sleep("5m30s")
sleep("1.1h2.2m3.3s")
`
	setupAndRunCode(t, script, "--color=never")
	assertSleptMillis(t, 10_000, 10_200, 12_345, 330_000, 4_095_300)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}

func TestSleep_HumanStringAllowSpaces(t *testing.T) {
	script := `
sleep("10 s")
sleep("10.2 s")
sleep("12345 ms")
sleep("5m 30s")
sleep("1.1h 2.2m  3.3s")
`
	setupAndRunCode(t, script, "--color=never")
	assertSleptMillis(t, 10_000, 10_200, 12_345, 330_000, 4_095_300)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}

func TestSleep_ErrorsIfNoArg(t *testing.T) {
	setupAndRunCode(t, `sleep()`, "--color=never")
	assertDidNotSleep(t)
	expected := `Error at L1:1

  sleep()
  ^^^^^^^ sleep() requires at least 1 argument, but got 0
`
	assertError(t, 1, expected)
}

func TestSleep_ErrorsIfNegArg(t *testing.T) {
	setupAndRunCode(t, `sleep(-10)`, "--color=never")
	assertDidNotSleep(t)
	expected := `Error at L1:7

  sleep(-10)
        ^^^ sleep() cannot take a negative duration: "-10s"
`
	assertError(t, 1, expected)
}

func TestSleep_ErrorsIfTooManyPositionalArgs(t *testing.T) {
	setupAndRunCode(t, `sleep(10, 20)`, "--color=never")
	assertDidNotSleep(t)
	expected := `Error at L1:1

  sleep(10, 20)
  ^^^^^^^^^^^^^ sleep() requires at most 1 argument, but got 2
`
	assertError(t, 1, expected)
}

func TestSleep_ErrorsIfIncorrectArgType(t *testing.T) {
	setupAndRunCode(t, `sleep(true)`, "--color=never")
	assertDidNotSleep(t)
	expected := `Error at L1:7

  sleep(true)
        ^^^^
        Got "bool" as the 1st argument of sleep(), but must be: int, float, or string
`
	assertError(t, 1, expected)
}

func TestSleep_ErrorsIfInvalidString(t *testing.T) {
	setupAndRunCode(t, `sleep("Invalid!")`, "--color=never")
	assertDidNotSleep(t)
	expected := `Error at L1:7

  sleep("Invalid!")
        ^^^^^^^^^^ Invalid string argument: "Invalid!"
`
	assertError(t, 1, expected)
}

func TestSleep_CanSleepLessThanMilliWithoutErroring(t *testing.T) {
	setupAndRunCode(t, `sleep(0.0001)`)
	assertSleptMillis(t, 0)
	assertAllElseEmpty(t)
	assertNoErrors(t)
}
