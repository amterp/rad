package testing

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestSnapshots runs snapshot tests from the snapshots/ directory.
// These tests support stdout, stderr, and exit code validation.
func TestSnapshots(t *testing.T) {
	snapshotDir := "snapshots"

	// Check if the directory exists
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		t.Skip("snapshots directory does not exist yet")
		return
	}

	runSnapshotDirectory(t, snapshotDir)
}

// runSnapshotDirectory runs all snapshot tests in the given directory.
func runSnapshotDirectory(t *testing.T, snapshotDir string) {
	// Find all .snap files
	var snapFiles []string
	err := filepath.Walk(snapshotDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".snap") {
			snapFiles = append(snapFiles, path)
		}
		return nil
	})
	require.NoError(t, err, "Failed to walk snapshot directory")

	if len(snapFiles) == 0 {
		t.Skip("No snapshot files found")
		return
	}

	// Track which files need updating (thread-safe for consistency,
	// even though tests don't run in parallel due to shared global state)
	var updateMu sync.Mutex
	filesToUpdate := make(map[string][]SnapshotCase)

	for _, snapFile := range snapFiles {
		snapFile := snapFile // capture for closure

		cases, err := ParseSnapshotFile(snapFile)
		require.NoError(t, err, "Failed to parse snapshot file: %s", snapFile)

		for i := range cases {
			tc := &cases[i]
			testName := strings.TrimPrefix(snapFile, snapshotDir+"/")
			testName = strings.TrimSuffix(testName, ".snap")
			if tc.Title != "" {
				testName = testName + "/" + tc.Title
			}

			// Note: We don't use t.Parallel() here because we rely on global state
			// (stdInBuffer, stdOutBuffer, stdErrBuffer) defined in test_helpers.go.
			t.Run(testName, func(t *testing.T) {
				if tc.SkipReason != "" {
					t.Skip(tc.SkipReason)
				}

				runSnapshotTest(t, tc)

				actual := SnapshotResult{
					Stdout:   normalizeOutput(stdOutBuffer.String()),
					Stderr:   normalizeOutput(stdErrBuffer.String()),
					ExitCode: getExitCode(),
				}

				if CompareSnapshotResult(t, tc, actual) {
					updateMu.Lock()
					tc.Stdout = actual.Stdout
					tc.Stderr = actual.Stderr
					tc.ExitCode = actual.ExitCode
					filesToUpdate[snapFile] = cases
					updateMu.Unlock()
				}
			})
		}
	}

	// Write updates if in update mode
	if *UpdateSnapshots {
		for path, cases := range filesToUpdate {
			err := WriteSnapshotFile(path, cases)
			if err != nil {
				t.Errorf("Failed to update snapshot file %s: %v", path, err)
			} else {
				t.Logf("Updated snapshot file: %s", path)
			}
		}
	}
}

// runSnapshotTest runs a single snapshot test case.
func runSnapshotTest(t *testing.T, tc *SnapshotCase) {
	t.Helper()

	// Build args, only adding --color=never if:
	// - RawArgs is false (normal mode)
	// - No --color flag is already specified
	args := tc.Args
	if !tc.RawArgs {
		hasColorFlag := false
		for _, arg := range args {
			if strings.HasPrefix(arg, "--color") {
				hasColorFlag = true
				break
			}
		}
		if !hasColorFlag {
			args = append(args, "--color=never")
		}
	}

	// Use the standard test setup
	setupAndRunCode(t, tc.Input, args...)
}

// getExitCode returns the exit code from the last test run.
// Returns 0 if no exit occurred.
func getExitCode() int {
	if errorOrExit.exitCode != nil {
		return *errorOrExit.exitCode
	}
	return 0
}

// normalizeOutput replaces test-specific file names with <script> for
// consistent snapshot comparison across different test environments.
func normalizeOutput(output string) string {
	// The test framework uses "TestCase" as the filename
	// Replace it with <script> for portable snapshots
	output = strings.ReplaceAll(output, "--> TestCase:", "--> <script>:")
	return output
}
