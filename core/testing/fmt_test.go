package testing

import "testing"

// The unformatted-file list is data (like `gofmt -l`): it must go to stdout so
// it can be piped onward, with stderr reserved for diagnostics.
// https://github.com/amterp/rad/issues/145

func Test_Fmt_Check_UnformattedFileListedOnStdout(t *testing.T) {
	setupAndRunArgs(t, "fmt", "--check", "./rad_scripts/unformatted.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "./rad_scripts/unformatted.rad\n")
	assertExitCode(t, 1)
}

func Test_Fmt_Check_FormattedFileNoOutput(t *testing.T) {
	setupAndRunArgs(t, "fmt", "--check", "./rad_scripts/hello.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "")
	assertNoErrors(t)
}

func Test_Fmt_Check_MixedListsOnlyUnformatted(t *testing.T) {
	setupAndRunArgs(t, "fmt", "--check", "./rad_scripts/hello.rad", "./rad_scripts/unformatted.rad", "--color=never")
	assertOnlyOutput(t, stdOutBuffer, "./rad_scripts/unformatted.rad\n")
	assertExitCode(t, 1)
}

func Test_Fmt_Check_UnparseableErrorsOnStderr(t *testing.T) {
	setupAndRunArgs(t, "fmt", "--check", "./rad_scripts/invalid.rad", "--color=never")
	assertError(t, 1, "./rad_scripts/invalid.rad: could not parse, skipping.\n")
}
