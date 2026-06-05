package radfmt

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	gd "github.com/amterp/go-delta"
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

// collectSnapshotInputs reads the INPUT section of every .snap file under
// snapshots/ by lightly parsing the snapshot format (### INPUT ### up to the
// next ### delimiter). It avoids importing core/testing so this internal test
// has no dependency cycle concerns.
func collectSnapshotInputs(t *testing.T) []snapshotInput {
	t.Helper()
	var out []snapshotInput
	dir := "snapshots"
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return nil
	}
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".snap") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		base := strings.TrimSuffix(filepath.Base(path), ".snap")
		inputs := parseSnapInputs(string(data))
		for i, in := range inputs {
			name := base
			if len(inputs) > 1 {
				name = base + "#" + strconv.Itoa(i)
			}
			out = append(out, snapshotInput{name: name, src: in})
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk snapshots: %v", err)
	}
	return out
}

// parseSnapInputs extracts each ### INPUT ### block's contents from a snapshot
// file body. This deliberately hand-rolls a minimal parser rather than reusing
// core/testing.ParseSnapshotFile: this is a white-box (package radfmt) test that
// needs internal access to formatRaw/structuralDump, and importing core/testing
// here would form an import cycle (core/testing -> core -> rts/radfmt).
func parseSnapInputs(body string) []string {
	var inputs []string
	lines := strings.Split(body, "\n")
	title := ""
	i := 0
	for i < len(lines) {
		trimmed := strings.TrimSpace(lines[i])
		if trimmed == "### TITLE ###" && i+1 < len(lines) {
			title = lines[i+1]
			i += 2
			continue
		}
		if trimmed == "### INPUT ###" {
			i++
			var b strings.Builder
			for i < len(lines) && !strings.HasPrefix(strings.TrimSpace(lines[i]), "### ") {
				b.WriteString(lines[i])
				b.WriteByte('\n')
				i++
			}
			// Skip [raw] cases: their INPUT is a Go-quoted string (decoded only by
			// the snapshot harness), and byte-level whitespace rules don't change
			// node structure anyway, so there's nothing to assert here.
			if !strings.Contains(title, "[raw]") {
				inputs = append(inputs, strings.TrimSuffix(b.String(), "\n"))
			}
			// Each case carries its own title; clear it so an INPUT block missing
			// its TITLE header can't silently inherit a prior [raw] title and be
			// dropped from the structure check.
			title = ""
			continue
		}
		i++
	}
	return inputs
}
