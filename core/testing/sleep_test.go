package testing

import "testing"

func TestSleep_IntBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep(10)`)
	assertSleptMillis(t, 10000)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}

func TestSleep_FloatBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep(10.2)`)
	assertSleptMillis(t, 10200)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}

func TestSleep_IntInStringBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep("10")`)
	assertSleptMillis(t, 10000)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}

func TestSleep_FloatInStringBecomesSeconds(t *testing.T) {
	setupAndRunCode(t, `sleep("10.2")`)
	assertSleptMillis(t, 10200)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}

func TestSleep_AllowsZero(t *testing.T) {
	setupAndRunCode(t, `sleep(0)`)
	assertSleptMillis(t, 0)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}

func TestSleep_GoCompatibleHumanString(t *testing.T) {
	rsl := `
sleep("10s")
sleep("10.2s")
sleep("12345ms")
sleep("5m30s")
sleep("1.1h2.2m3.3s")
`
	setupAndRunCode(t, rsl)
	assertSleptMillis(t, 10_000, 10_200, 12_345, 330_000, 4_095_300)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}

func TestSleep_HumanStringAllowSpaces(t *testing.T) {
	rsl := `
sleep("10 s")
sleep("10.2 s")
sleep("12345 ms")
sleep("5m 30s")
sleep("1.1h 2.2m  3.3s")
`
	setupAndRunCode(t, rsl)
	assertSleptMillis(t, 10_000, 10_200, 12_345, 330_000, 4_095_300)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}

func TestSleep_ErrorsIfNoArg(t *testing.T) {
	setupAndRunCode(t, `sleep()`)
	assertDidNotSleep(t)
	assertError(t, 1, "RslError at L1/5 on 'sleep': sleep() takes exactly one positional argument\n")
	resetTestState()
}

func TestSleep_ErrorsIfNegArg(t *testing.T) {
	setupAndRunCode(t, `sleep(-10)`)
	assertDidNotSleep(t)
	assertError(t, 1, "RslError at L1/5 on 'sleep': sleep() cannot take a negative duration: -10s\n")
	resetTestState()
}

func TestSleep_ErrorsIfTooManyPositionalArgs(t *testing.T) {
	setupAndRunCode(t, `sleep(10, 20)`)
	assertDidNotSleep(t)
	assertError(t, 1, "RslError at L1/5 on 'sleep': sleep() takes exactly one positional argument\n")
	resetTestState()
}

func TestSleep_ErrorsIfIncorrectArgType(t *testing.T) {
	setupAndRunCode(t, `sleep(true)`)
	assertDidNotSleep(t)
	assertError(t, 1, "RslError at L1/5 on 'sleep': sleep() takes an int, float, or string, got bool\n")
	resetTestState()
}

func TestSleep_ErrorsIfInvalidString(t *testing.T) {
	setupAndRunCode(t, `sleep("Invalid!")`)
	assertDidNotSleep(t)
	assertError(t, 1, "RslError at L1/5 on 'sleep': invalid string argument: 'Invalid!'\n")
	resetTestState()
}

func TestSleep_CanSleepLessThanMilliWithoutErroring(t *testing.T) {
	setupAndRunCode(t, `sleep(0.0001)`)
	assertSleptMillis(t, 0)
	assertAllElseEmpty(t)
	assertNoErrors(t)
	resetTestState()
}
