package testing

import (
	"testing"
)

func TestPrint(t *testing.T) {
	setupAndRun(t, "./test_rads/print.rad")

	expected := `hi alice
hi bob
hi charlie
`
	assertOnly(t, stdOutBuffer, expected)
}
