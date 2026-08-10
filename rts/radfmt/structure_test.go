package radfmt

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	gd "github.com/amterp/go-delta"
	snap "github.com/amterp/go-snap"

	"github.com/amterp/rad/rts"
)

// dumpStructure parses src and returns its readable, position-free node-structure
// dump. Used to assert the formatter never changes what the code parses to.
func dumpStructure(t *testing.T, src string) string {
	t.Helper()
	parser, err := rts.NewRadParser()
	if err != nil {
		t.Fatalf("parser: %v", err)
	}
	defer parser.Close()
	tree := parser.Parse(src)
	return structuralDump(tree.Root())
}

// TestStructurePreserved is the explicit version of Format's runtime guard: for
// every snapshot input, it formats the source (raw, BEFORE the structural
// no-op) and asserts the CST node structure - kinds and field names, ignoring
// whitespace, positions, quote characters, and comment placement - is identical
// before and after. A formatter bug that reorders or drops nodes fails here with
// a readable side-by-side diff, rather than silently degrading to a no-op.
func TestStructurePreserved(t *testing.T) {
	inputs := collectSnapshotInputs(t)
	if len(inputs) == 0 {
		t.Skip("no snapshot inputs yet")
	}
	for _, in := range inputs {
		in := in
		t.Run(in.name, func(t *testing.T) {
			t.Parallel()
			raw, _, _, ok := formatRaw(normalizeLineEndings(in.src))
			if !ok {
				t.Skipf("input has parse errors: %s", in.name)
			}
			before := dumpStructure(t, normalizeLineEndings(in.src))
			after := dumpStructure(t, raw)
			if before != after {
				t.Errorf("formatting changed node structure for %s:\n%s",
					in.name,
					gd.DiffWith(before, after,
						gd.WithColor(true),
						gd.WithLayout(gd.LayoutPreferSideBySide),
						gd.WithWidth(120)))
			}
		})
	}
}

type snapshotInput struct {
	name string
	src  string
}

// inputSchema is the read-only view of the snapshot format this test needs. It
// is a separate declaration from radfmt_test.FmtSuite because this is a
// white-box test - it lives in package radfmt for access to formatRaw and
// structuralDump, and so cannot see the external test package's variables.
var inputSchema = snap.Suite{
	Inputs:  []snap.Input{{Name: "INPUT"}},
	Outputs: []snap.Output{{Name: "FORMATTED"}},
}

// collectSnapshotInputs reads the INPUT section of every case under snapshots/.
//
// This used to hand-roll a minimal parser, because importing rad's own snapshot
// format layer would have formed an import cycle (core/testing -> core ->
// rts/radfmt). go-snap is a separate module that imports nothing from rad, so
// the real parser is reachable now.
func collectSnapshotInputs(t *testing.T) []snapshotInput {
	t.Helper()
	dir := "snapshots"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	var out []snapshotInput
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".snap") {
			return err
		}
		cases, err := snap.ParseFile(&inputSchema, path)
		if err != nil {
			return err
		}
		base := strings.TrimSuffix(filepath.Base(path), ".snap")
		for i, c := range cases {
			name := base
			if len(cases) > 1 {
				name = base + "#" + strconv.Itoa(i)
			}
			out = append(out, snapshotInput{name: name, src: c.Text("INPUT")})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk snapshots: %v", err)
	}
	return out
}
