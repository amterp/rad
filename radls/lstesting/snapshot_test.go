package lstesting

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	gd "github.com/amterp/go-delta"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/amterp/rad/radls/log"
)

func TestMain(m *testing.M) {
	log.L = zap.NewNop().Sugar()
	os.Exit(m.Run())
}

func TestSnapshots(t *testing.T) {
	snapshotDir := "snapshots"

	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		t.Skip("snapshots directory does not exist yet")
		return
	}

	runSnapshotDirectory(t, snapshotDir)
}

func runSnapshotDirectory(t *testing.T, snapshotDir string) {
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

	var updateMu sync.Mutex
	filesToUpdate := make(map[string][]SnapshotCase)

	for _, snapFile := range snapFiles {
		snapFile := snapFile

		cases, err := ParseSnapshotFile(snapFile)
		require.NoError(t, err, "Failed to parse snapshot file: %s", snapFile)

		for i := range cases {
			tc := &cases[i]
			testName := strings.TrimPrefix(snapFile, snapshotDir+"/")
			testName = strings.TrimSuffix(testName, ".snap")
			if tc.Title != "" {
				testName = testName + "/" + tc.Title
			}

			t.Run(testName, func(t *testing.T) {
				t.Parallel()

				actual, err := Run(tc)
				require.NoError(t, err, "Harness should run without error")

				if actual != tc.Stdout {
					if *UpdateSnapshots {
						updateMu.Lock()
						tc.Stdout = actual
						filesToUpdate[snapFile] = cases
						updateMu.Unlock()
					} else {
						t.Errorf("Output mismatch:\n%s",
							gd.DiffWith(tc.Stdout, actual,
								gd.WithColor(true),
								gd.WithLayout(gd.LayoutPreferSideBySide),
								gd.WithWidth(120)))
					}
				}
			})
		}
	}

	// Write updates after all subtests complete
	if *UpdateSnapshots {
		t.Cleanup(func() {
			for path, cases := range filesToUpdate {
				err := WriteSnapshotFile(path, cases)
				if err != nil {
					t.Errorf("Failed to update snapshot file %s: %v", path, err)
				} else {
					t.Logf("Updated snapshot file: %s", path)
				}
			}
		})
	}
}
