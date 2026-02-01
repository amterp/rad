package testing

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestErrorSnapshots(t *testing.T) {
	snapshotDir := "error_snapshots"

	// Check if the directory exists
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		t.Skip("error_snapshots directory does not exist yet")
		return
	}

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

	// Track which files need updating (thread-safe for consistency with CST tests,
	// even though these tests don't run in parallel due to shared global state in test_helpers.go)
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

			// Note: We don't use t.Parallel() here because runErrorSnapshotTest
			// relies on global state (stdInBuffer, stdOutBuffer, stdErrBuffer)
			// defined in test_helpers.go. Parallelizing would cause data races.
			t.Run(testName, func(t *testing.T) {
				runErrorSnapshotTest(t, tc)

				// Get the actual stderr output and normalize it
				actual := normalizeOutput(stdErrBuffer.String())

				if CompareSnapshot(t, tc, actual) {
					// Needs update - update tc.Expected under lock
					updateMu.Lock()
					tc.Expected = actual
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

func runErrorSnapshotTest(t *testing.T, tc *SnapshotCase) {
	t.Helper()

	// Use the standard test setup which handles colors via --color=never
	setupAndRunCode(t, tc.Input, "--color=never")
}

// normalizeOutput replaces test-specific file names with <script> for
// consistent snapshot comparison across different test environments
func normalizeOutput(output string) string {
	// The test framework uses "TestCase" as the filename
	// Replace it with <script> for portable snapshots
	output = strings.ReplaceAll(output, "--> TestCase:", "--> <script>:")
	return output
}
