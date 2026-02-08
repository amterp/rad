package rts_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	radtesting "github.com/amterp/rad/core/testing"
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/require"
)

func TestSTSnapshots(t *testing.T) {
	snapshotDir := "test/st_snapshots"

	// Check if the directory exists
	if _, err := os.Stat(snapshotDir); os.IsNotExist(err) {
		t.Skip("st_snapshots directory does not exist yet")
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

	// Track which files need updating (thread-safe since tests run in parallel)
	var updateMu sync.Mutex
	filesToUpdate := make(map[string][]radtesting.SnapshotCase)

	for _, snapFile := range snapFiles {
		snapFile := snapFile // capture for closure

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

				parser, err := rts.NewRadParser()
				require.NoError(t, err, "Failed to create parser")
				defer parser.Close()

				tree := parser.Parse(tc.Input)
				cstDump := tree.Dump()

				// Only generate AST if there are no parse errors
				// Recover from panics during conversion (e.g., out-of-range numbers)
				astDump := ""
				if !tree.Root().HasError() {
					func() {
						defer func() {
							if r := recover(); r != nil {
								// AST conversion failed, leave astDump empty
								t.Logf("AST conversion failed: %v", r)
							}
						}()
						ast := rts.ConvertCST(tree.Root(), tc.Input, "test.rad")
						astDump = rl.AstDump(ast)
					}()
				}

				actual := radtesting.SnapshotResult{
					Stdout: cstDump,
					Stderr: astDump,
				}

				if radtesting.CompareSnapshotResult(t, tc, actual) {
					updateMu.Lock()
					tc.Stdout = actual.Stdout
					tc.Stderr = actual.Stderr
					filesToUpdate[snapFile] = cases
					updateMu.Unlock()
				}
			})
		}
	}

	// Write updates after all subtests complete
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
