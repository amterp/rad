package check_test

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	gd "github.com/amterp/go-delta"
	radtesting "github.com/amterp/rad/core/testing"
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/require"
	tsapi "github.com/tree-sitter/go-tree-sitter"
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

				// Run the full Check() pipeline (binder + type
				// checker + AST-level checks) rather than just
				// Resolve + TypeCheck. Without the AST checks,
				// snapshots can claim "no diagnostics" while
				// `rad check` emits real errors (return-outside-
				// fn, invalid LHS, etc.).
				parser, err := rts.NewRadParser()
				require.NoError(t, err)
				defer parser.Close()
				tree := parser.Parse(tc.Input)
				// Match the production check pipeline's recover
				// guard: malformed inputs (the kind that exercise
				// parser-error heuristics like RAD10009) produce
				// ERROR nodes the converter doesn't know how to
				// represent. The CLI's `rad check` recovers and
				// proceeds without an AST; snapshots should too.
				file := safeConvertCST(tree.Root(), tc.Input)
				checker := check.NewCheckerWithTree(tree, parser, tc.Input, file)
				result, err := checker.Check()
				require.NoError(t, err, "Check should not error")
				// Resolved / Types may be nil when the converter
				// bailed on malformed input; DumpForSnapshot
				// tolerates that (parser-error heuristic
				// snapshots exercise the nil-ast path).

				actual := check.DumpForSnapshot(file, result.Types, result.Resolved, result.Diagnostics)

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

// safeConvertCST wraps rts.ConvertCST with a panic recover, matching
// the production check pipeline (rts/check/check.go tryConvertAST).
// Returns nil when the converter bails on malformed input; the
// checker tolerates a nil ast in that case.
func safeConvertCST(root *tsapi.Node, src string) (file *rl.SourceFile) {
	defer func() {
		if r := recover(); r != nil {
			file = nil
		}
	}()
	return rts.ConvertCST(root, src, "snapshot_test.rad")
}
