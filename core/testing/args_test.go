package testing

import "testing"

const (
	setupArgRsl = `
args:
    foo "bar" x string`
)

func TestArgApiRename(t *testing.T) {
	rsl := setupArgRsl + `
print(foo)
`
	setupAndRunCode(t, rsl, "hey")
	expected := `hey
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}

func TestArgApiRenameUsageString(t *testing.T) {
	setupAndRunCode(t, setupArgRsl, "-h")
	expected := `Usage:
  test <bar> [flags]

Flags:
  -x, --bar string
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
	resetTestState()
}
