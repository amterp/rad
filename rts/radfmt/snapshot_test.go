package radfmt_test

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"

	gd "github.com/amterp/go-delta"
	radtesting "github.com/amterp/rad/core/testing"
	"github.com/amterp/rad/rts/radfmt"
	"github.com/stretchr/testify/require"
)

// rawMarker in a snapshot title marks a byte-level case whose INPUT and STDOUT
// are Go-quoted strings (e.g. "x = 1\r\n"). Line-ending, trailing-whitespace,
// and exact-trailing-newline rules can't be represented as literal snapshot text
// - the scanner strips CRs and editors strip trailing spaces - so these cases
// encode the bytes as visible escapes and are compared byte-for-byte.
const rawMarker = "[raw]"

func unquoteSnap(t *testing.T, s string) string {
	v, err := strconv.Unquote(s)
	require.NoErrorf(t, err, "[raw] snapshot section must be a Go-quoted string, got: %s", s)
	return v
}

// TestFmtSnapshots runs every .snap file under rts/radfmt/snapshots/ through
// Format and compares the result against the snapshot's STDOUT section. The
// INPUT is intentionally-messy Rad source; STDOUT is the canonical formatting.
//
// To regenerate expected outputs after intentional formatter changes:
//
//	go test ./rts/radfmt/ -run TestFmtSnapshots -update
//
// Because ParseSnapshotFile strips the trailing newline from both sections,
// comparisons normalize the trailing newline away; the exact EOF-newline rule
// is covered by unit tests in fmt_test.go.
func TestFmtSnapshots(t *testing.T) {
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

				// A [raw] case carries its input and expected output as Go-quoted
				// strings; decode them and compare byte-for-byte (no trailing-
				// newline trim) so byte-level rules are precisely testable.
				raw := strings.Contains(tc.Title, rawMarker)
				input, expected := tc.Input, tc.Stdout
				if raw {
					input = unquoteSnap(t, tc.Input)
					expected = unquoteSnap(t, tc.Stdout)
				}

				out, _, ok := radfmt.Format(input)
				require.True(t, ok, "Format returned ok=false (parse error?) for input:\n%s", input)

				actual := out
				if !raw {
					actual = strings.TrimRight(out, "\n")
				}

				if actual != expected {
					if *radtesting.UpdateSnapshots {
						updateMu.Lock()
						if raw {
							tc.Stdout = strconv.Quote(actual)
						} else {
							tc.Stdout = actual
						}
						filesToUpdate[snapFile] = cases
						updateMu.Unlock()
					} else if raw {
						t.Errorf("Snapshot mismatch for %s:\n expected %s\n   actual %s",
							tc.Title, strconv.Quote(expected), strconv.Quote(actual))
					} else {
						t.Errorf("Snapshot mismatch for %s:\n%s",
							tc.Title,
							gd.DiffWith(expected, actual,
								gd.WithColor(true),
								gd.WithLayout(gd.LayoutPreferSideBySide),
								gd.WithWidth(120)))
					}
				}

				// Idempotence: formatting already-formatted output is a no-op.
				reformatted, changed, ok2 := radfmt.Format(out)
				require.True(t, ok2, "re-format returned ok=false")
				require.Equal(t, out, reformatted, "format is not idempotent")
				require.False(t, changed, "re-format reported a change (not idempotent)")
			})
		}
	}

	t.Cleanup(func() {
		if *radtesting.UpdateSnapshots {
			for path, cases := range filesToUpdate {
				if err := radtesting.WriteSnapshotFile(path, cases); err != nil {
					t.Errorf("Failed to update snapshot file %s: %v", path, err)
				} else {
					t.Logf("Updated snapshot file: %s", path)
				}
			}
		}
	})
}
