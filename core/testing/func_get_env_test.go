package testing

import (
	"os"
	"testing"
)

func Test_GetEnv_ExistingVar(t *testing.T) {
	os.Setenv("TEST_ENV_VAR", "test_value")
	defer os.Unsetenv("TEST_ENV_VAR")

	script := `
print(get_env("TEST_ENV_VAR"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "test_value\n")
	assertNoErrors(t)
}

func Test_GetEnv_NonExistentVar(t *testing.T) {
	os.Unsetenv("NONEXISTENT_ENV_VAR")

	script := `
print(get_env("NONEXISTENT_ENV_VAR"))
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "\n")
	assertNoErrors(t)
}

func Test_GetEnv_NonExistentVarCanOr(t *testing.T) {
	os.Unsetenv("NONEXISTENT_ENV_VAR")

	script := `
print(get_env("NONEXISTENT_ENV_VAR") or "bloopy")
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "bloopy\n")
	assertNoErrors(t)
}

func Test_GetEnv_Assignment(t *testing.T) {
	os.Setenv("ANOTHER_TEST_VAR", "another_value")
	defer os.Unsetenv("ANOTHER_TEST_VAR")

	script := `
env_value = get_env("ANOTHER_TEST_VAR")
print(env_value)
`
	setupAndRunCode(t, script, "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "another_value\n")
	assertNoErrors(t)
}
