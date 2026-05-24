package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

func documentSymbolsFixture(t *testing.T, src string) []lsp.DocumentSymbol {
	t.Helper()
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///doc_test.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	defer snap.Release()
	syms, err := s.DocumentSymbols(snap)
	if err != nil {
		t.Fatalf("DocumentSymbols: %v", err)
	}
	return syms
}

// TestDocumentSymbolsEmptyDocument verifies we return an empty slice
// (not nil) for an empty document, so the JSON-RPC reply is `[]` and
// not `null`. Some clients reject the latter.
func TestDocumentSymbolsEmptyDocument(t *testing.T) {
	syms := documentSymbolsFixture(t, "")
	if syms == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols, got %d", len(syms))
	}
}

// TestDocumentSymbolsTopLevelAssignments verifies top-level
// `x = ...` assignments produce one Variable each, in source order,
// with a SelectionRange covering just the name.
func TestDocumentSymbolsTopLevelAssignments(t *testing.T) {
	syms := documentSymbolsFixture(t, "x = 1\ny = 2\n")
	if len(syms) != 2 {
		t.Fatalf("expected 2 symbols, got %d (%v)", len(syms), syms)
	}
	if syms[0].Name != "x" || syms[1].Name != "y" {
		t.Errorf("names: got [%s %s], want [x y]", syms[0].Name, syms[1].Name)
	}
	for i, sym := range syms {
		if sym.Kind != lsp.SymbolKindVariable {
			t.Errorf("symbol[%d]: kind=%d, want %d", i, sym.Kind, lsp.SymbolKindVariable)
		}
	}
}

// TestDocumentSymbolsReassignmentDoesntDuplicate verifies a second
// `x = 2` doesn't appear as a separate outline entry. Duplicate
// entries make the outline noisy without adding signal.
func TestDocumentSymbolsReassignmentDoesntDuplicate(t *testing.T) {
	syms := documentSymbolsFixture(t, "x = 1\nx = 2\n")
	if len(syms) != 1 {
		t.Errorf("expected 1 symbol for x (only first decl), got %d (%v)",
			len(syms), syms)
	}
}

// TestDocumentSymbolsFunctionDef verifies a top-level fn produces a
// Function symbol whose SelectionRange covers just the name and
// whose Range covers the whole def.
func TestDocumentSymbolsFunctionDef(t *testing.T) {
	src := "fn greet():\n    print(\"hi\")\n"
	syms := documentSymbolsFixture(t, src)
	if len(syms) != 1 {
		t.Fatalf("expected 1 symbol, got %d (%v)", len(syms), syms)
	}
	if syms[0].Name != "greet" {
		t.Errorf("name: got %q, want greet", syms[0].Name)
	}
	if syms[0].Kind != lsp.SymbolKindFunction {
		t.Errorf("kind: got %d, want %d", syms[0].Kind, lsp.SymbolKindFunction)
	}
	// Range should cover whole def (line 0 -> at least line 1);
	// SelectionRange should be narrower (just name).
	if syms[0].Range.End.Line < 1 {
		t.Errorf("function Range should span body: got %+v", syms[0].Range)
	}
	if syms[0].SelectionRange.End.Line != 0 {
		t.Errorf("function SelectionRange should be name-only: got %+v",
			syms[0].SelectionRange)
	}
}

// TestDocumentSymbolsArgsBlock verifies the `args:` block becomes a
// Namespace named "args" with one Variable child per declared arg.
func TestDocumentSymbolsArgsBlock(t *testing.T) {
	src := `args:
    name str
    age int = 30
`
	syms := documentSymbolsFixture(t, src)
	if len(syms) != 1 {
		t.Fatalf("expected 1 top-level symbol, got %d (%v)", len(syms), syms)
	}
	if syms[0].Name != "args" || syms[0].Kind != lsp.SymbolKindNamespace {
		t.Errorf("args block symbol: %+v", syms[0])
	}
	if len(syms[0].Children) != 2 {
		t.Fatalf("expected 2 children, got %d (%v)",
			len(syms[0].Children), syms[0].Children)
	}
	childNames := []string{syms[0].Children[0].Name, syms[0].Children[1].Name}
	if childNames[0] != "name" || childNames[1] != "age" {
		t.Errorf("child names: got %v, want [name age]", childNames)
	}
	// Detail should carry the type. Spot-check on the second child.
	if syms[0].Children[1].Detail == "" {
		t.Errorf("expected Detail for age, got empty")
	}
}

// TestDocumentSymbolsInSourceOrder verifies the outline matches
// the file's actual order. Before the sort fix, args:/cmd:
// always appeared first regardless of where they sat in source,
// surprising users scanning the panel for an entry near a
// specific line.
func TestDocumentSymbolsInSourceOrder(t *testing.T) {
	// args block, then variable, then function - all in that
	// order textually. Without the sort fix, traversal order
	// happened to match by coincidence; with the sort it's
	// guaranteed.
	src := `args:
    name str

x = 1

fn helper():
    print(1)
`
	syms := documentSymbolsFixture(t, src)
	if len(syms) != 3 {
		t.Fatalf("expected 3 symbols, got %d (%v)", len(syms), syms)
	}
	names := []string{syms[0].Name, syms[1].Name, syms[2].Name}
	want := []string{"args", "x", "helper"}
	for i, n := range names {
		if n != want[i] {
			t.Errorf("symbol[%d]: got %q, want %q (source order)",
				i, n, want[i])
		}
	}
	// And the lines should be ascending.
	for i := 1; i < len(syms); i++ {
		if syms[i].Range.Start.Line < syms[i-1].Range.Start.Line {
			t.Errorf("symbols not in line order: %d before %d",
				syms[i-1].Range.Start.Line, syms[i].Range.Start.Line)
		}
	}
}

// TestDocumentSymbolsNoSnapshotReturnsEmpty verifies the nil path.
func TestDocumentSymbolsNoSnapshotReturnsEmpty(t *testing.T) {
	s := NewState()
	syms, err := s.DocumentSymbols(nil)
	if err != nil {
		t.Fatalf("DocumentSymbols: %v", err)
	}
	if syms == nil {
		t.Errorf("expected empty slice for nil snapshot, got nil")
	}
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols for nil snapshot, got %d", len(syms))
	}
}
