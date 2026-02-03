package testing

import "testing"

// KEEP_GO: Uses setupAndRunArgs with script file path - tests that resource
// paths are resolved relative to the script's location, not cwd. Cannot be
// migrated to snapshot tests which run inline scripts from temp files.
func Test_ResourcePathIsRelativeToScript(t *testing.T) {
	setupAndRunArgs(t, "./rad_scripts/people_resource.rad", "bob")
	expected := `Bob
350
`
	assertOnlyOutput(t, stdOutBuffer, expected)
	assertNoErrors(t)
}
