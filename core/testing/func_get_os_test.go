package testing

import (
	"runtime"
	"testing"
)

func Test_Func_GetOs(t *testing.T) {
	script := `
print(get_os())
`
	setupAndRunCode(t, script, "--color=never")

	var expected string
	switch runtime.GOOS {
	case "darwin":
		expected = "macos\n"
	default:
		expected = runtime.GOOS + "\n"
	}
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
