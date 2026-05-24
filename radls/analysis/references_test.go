package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

func referencesFixture(t *testing.T, src string, pos lsp.Pos, includeDecl bool) []lsp.Location {
	t.Helper()
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///refs_test.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	defer snap.Release()
	locs, err := s.References(snap, pos, includeDecl)
	if err != nil {
		t.Fatalf("References: %v", err)
	}
	return locs
}

// TestReferencesUseSitesOnly verifies includeDeclaration=false
// returns just the use sites of a symbol. Three references of `x`
// past the decl -> three Locations, no decl.
func TestReferencesUseSitesOnly(t *testing.T) {
	src := "x = 1\nprint(x)\nprint(x + x)\n"
	// Cursor on the `x` on line 0 (the decl).
	locs := referencesFixture(t, src, lsp.NewPos(0, 0), false)
	if len(locs) != 3 {
		t.Errorf("expected 3 uses (excluding decl), got %d (%v)", len(locs), locs)
	}
}

// TestReferencesIncludesDecl verifies includeDeclaration=true adds
// the decl site to the result.
func TestReferencesIncludesDecl(t *testing.T) {
	src := "x = 1\nprint(x)\n"
	locs := referencesFixture(t, src, lsp.NewPos(0, 0), true)
	if len(locs) != 2 {
		t.Errorf("expected 2 locations (decl + 1 use), got %d (%v)",
			len(locs), locs)
	}
}

// TestReferencesSortedByPosition verifies results come back in
// source order. Map iteration in Go is non-deterministic; without
// the explicit sort, the editor would show results in whatever
// order the underlying map happened to walk this time.
func TestReferencesSortedByPosition(t *testing.T) {
	src := "x = 1\nprint(x)\nprint(x + x)\n"
	locs := referencesFixture(t, src, lsp.NewPos(0, 0), false)
	if len(locs) < 2 {
		t.Fatalf("need >= 2 to test sort, got %d", len(locs))
	}
	for i := 1; i < len(locs); i++ {
		prev, cur := locs[i-1].Range.Start, locs[i].Range.Start
		if prev.Line > cur.Line ||
			(prev.Line == cur.Line && prev.Character > cur.Character) {
			t.Errorf("not sorted: %v before %v", prev, cur)
		}
	}
}

// TestReferencesFromUseSite verifies the cursor on a use (not the
// decl) still finds all references. The lookup goes through the
// Symbol, so any identifier bound to that Symbol must produce the
// same result set.
func TestReferencesFromUseSite(t *testing.T) {
	src := "x = 1\nprint(x)\nprint(x)\n"
	// Cursor on the `x` in `print(x)` on line 1.
	locs := referencesFixture(t, src, lsp.NewPos(1, 6), false)
	if len(locs) != 2 {
		t.Errorf("expected 2 uses from use-site cursor, got %d (%v)",
			len(locs), locs)
	}
}

// TestReferencesOffIdentifierReturnsEmpty verifies cursor on
// non-identifier territory yields an empty array (not nil).
func TestReferencesOffIdentifierReturnsEmpty(t *testing.T) {
	locs := referencesFixture(t, "x = 1\n", lsp.NewPos(0, 4), false)
	if locs == nil {
		t.Fatal("expected empty slice, got nil")
	}
	if len(locs) != 0 {
		t.Errorf("expected 0 locations, got %d (%v)", len(locs), locs)
	}
}

// TestReferencesNoSnapshotReturnsEmpty verifies the nil-snapshot path.
func TestReferencesNoSnapshotReturnsEmpty(t *testing.T) {
	s := NewState()
	locs, err := s.References(nil, lsp.NewPos(0, 0), true)
	if err != nil {
		t.Fatalf("References: %v", err)
	}
	if locs == nil || len(locs) != 0 {
		t.Errorf("expected empty slice for nil snapshot, got %v", locs)
	}
}

// TestReferencesURIMatchesDocument verifies every Location carries
// the document's URI. Single-file project today; this is forward
// insurance for when multi-file lookups arrive.
func TestReferencesURIMatchesDocument(t *testing.T) {
	src := "x = 1\nprint(x)\n"
	locs := referencesFixture(t, src, lsp.NewPos(0, 0), true)
	if len(locs) == 0 {
		t.Fatal("expected non-empty")
	}
	for i, loc := range locs {
		if loc.Uri != "file:///refs_test.rad" {
			t.Errorf("loc[%d] uri: got %q, want %q",
				i, loc.Uri, "file:///refs_test.rad")
		}
	}
}
