package check_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	gd "github.com/amterp/go-delta"
	radtesting "github.com/amterp/rad/core/testing"
	"github.com/amterp/rad/rts/check"
	"github.com/stretchr/testify/require"
)

// TestCheckSnapshots runs every .snap file under rts/check/snapshots/
// through the binder + type checker and compares the deterministic
// dump against the snapshot's expected output.
//
// To regenerate snapshots after intentional behavior changes:
//
//	go test ./rts/check/ -run TestCheckSnapshots -update
//
// The harness is intentionally simple: each case'\''s INPUT is a Rad
// script, the STDOUT section is the expected DumpForSnapshot output.
// Stderr / exit-code aren'\''t used (the checker isn'\''t a runtime).
func TestCheckSnapshots(t *testing.T) {
	snapshotDir := "snapshots"
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		t.Skip("snapshots directory does not exist yet")
		return
	}

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
	filesToUpdate := make(map[string][]radtesting.SnapshotCase)

	for _, snapFile := range snapFiles {
		snapFile := snapFile

		cases, err := radtesting.ParseSnapshotFile(snapFile)
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
				if tc.SkipReason != "" {
					t.Skip(tc.SkipReason)
				}

				file := parseFile(t, tc.Input)
				resolved := check.Resolve(file)
				require.NotNil(t, resolved, "Resolve should not return nil")
				info := check.TypeCheck(file, resolved)
				require.NotNil(t, info, "TypeCheck should not return nil")

				actual := check.DumpForSnapshot(file, info, resolved)

				if actual != tc.Stdout {
					if *radtesting.UpdateSnapshots {
						updateMu.Lock()
						tc.Stdout = actual
						filesToUpdate[snapFile] = cases
						updateMu.Unlock()
					} else {
						t.Errorf("Snapshot mismatch for %s:\n%s",
							tc.Title,
							gd.DiffWith(tc.Stdout, actual,
								gd.WithColor(true),
								gd.WithLayout(gd.LayoutPreferSideBySide),
								gd.WithWidth(120)))
					}
				}
			})
		}
	}

	t.Cleanup(func() {
		if *radtesting.UpdateSnapshots {
			for path, cases := range filesToUpdate {
				err := radtesting.WriteSnapshotFile(path, cases)
				if err != nil {
					t.Errorf("Failed to update snapshot file %s: %v", path, err)
				} else {
					t.Logf("Updated snapshot file: %s", path)
				}
			}
		}
	})
}
