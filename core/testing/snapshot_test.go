package testing

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const snapshotDelimiter = "### EXPECTED ###"

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

	for _, snapFile := range snapFiles {
		// Use relative path for test name
		testName := strings.TrimPrefix(snapFile, snapshotDir+"/")
		testName = strings.TrimSuffix(testName, ".snap")

		t.Run(testName, func(t *testing.T) {
			runSnapshotTest(t, snapFile)
		})
	}
}

func runSnapshotTest(t *testing.T, snapFile string) {
	t.Helper()

	// Read the snapshot file
	content, err := os.ReadFile(snapFile)
	require.NoError(t, err, "Failed to read snapshot file: %s", snapFile)

	// Split by delimiter
	parts := strings.SplitN(string(content), snapshotDelimiter, 2)
	if len(parts) != 2 {
		t.Fatalf("Snapshot file %s missing delimiter '%s'", snapFile, snapshotDelimiter)
	}

	script := strings.TrimSpace(parts[0])
	expected := strings.TrimPrefix(parts[1], "\n") // Remove leading newline after delimiter

	// Use the standard test setup which handles colors via --color=never
	setupAndRunCode(t, script, "--color=never")

	// Get the actual stderr output and normalize it
	actual := normalizeOutput(stdErrBuffer.String())

	// Compare
	if !assert.Equal(t, expected, actual, "Snapshot mismatch in %s", snapFile) {
		// Print a helpful diff
		t.Logf("Expected:\n%s", expected)
		t.Logf("Actual:\n%s", actual)
	}
}

// normalizeOutput replaces test-specific file names with <script> for
// consistent snapshot comparison across different test environments
func normalizeOutput(output string) string {
	// The test framework uses "TestCase" as the filename
	// Replace it with <script> for portable snapshots
	output = strings.ReplaceAll(output, "--> TestCase:", "--> <script>:")
	return output
}
