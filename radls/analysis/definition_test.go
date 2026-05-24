package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

func definitionFixture(t *testing.T, src string, pos lsp.Pos) *lsp.Location {
	t.Helper()
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///def_test.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot after AddDoc")
	}
	defer snap.Release()
	loc, err := s.Definition(snap, pos)
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	return loc
}

// TestDefinitionLocalUseJumpsToDecl exercises the canonical case:
// reference -> declaration on the same line is one Location with the
// LHS span. This is the call most users make when they hit "go to
// definition" on a variable.
func TestDefinitionLocalUseJumpsToDecl(t *testing.T) {
	// `x = 1\nprint(x)` - cursor on the `x` argument inside print()
	// at line 1 col 6. The decl is `x` at line 0 col 0.
	src := "x = 1\nprint(x)\n"
	loc := definitionFixture(t, src, lsp.NewPos(1, 6))
	if loc == nil {
		t.Fatal("expected location for local use, got nil")
	}
	if loc.Range.Start.Line != 0 || loc.Range.Start.Character != 0 {
		t.Errorf("start: got %+v, want (0,0)", loc.Range.Start)
	}
	// `x` is one char wide.
	if loc.Range.End.Character != 1 {
		t.Errorf("end character: got %d, want 1", loc.Range.End.Character)
	}
}

// TestDefinitionBuiltinReturnsNil verifies "go to def" on a builtin
// gives no location rather than e.g. (0,0). Builtins have no source
// span; the hover already shows their signature, so navigating
// somewhere arbitrary would be a poor UX.
func TestDefinitionBuiltinReturnsNil(t *testing.T) {
	// Cursor on `print` (a builtin) at line 0 col 0.
	loc := definitionFixture(t, "print(1)\n", lsp.NewPos(0, 0))
	if loc != nil {
		t.Errorf("expected nil location for builtin, got %+v", loc)
	}
}

// TestDefinitionOffIdentifierReturnsNil verifies that hovering off
// any identifier (whitespace, punctuation) returns nil. Editors
// treat null as "feature is available but not here," which is the
// right experience.
func TestDefinitionOffIdentifierReturnsNil(t *testing.T) {
	loc := definitionFixture(t, "x = 1   \n", lsp.NewPos(0, 7))
	if loc != nil {
		t.Errorf("expected nil off-identifier, got %+v", loc)
	}
}

// TestDefinitionFunctionCallJumpsToFnDef verifies that calling a
// user-defined function navigates to its `fn` keyword span. This
// is the second-most-common navigation, after local-variable
// lookup.
func TestDefinitionFunctionCallJumpsToFnDef(t *testing.T) {
	// Hoisted-fn binding spans the entire FnDef per the binder.
	src := "fn greet():\n    print(\"hi\")\n\ngreet()\n"
	loc := definitionFixture(t, src, lsp.NewPos(3, 0))
	if loc == nil {
		t.Fatal("expected location for user-fn call, got nil")
	}
	if loc.Range.Start.Line != 0 {
		t.Errorf("expected decl on line 0, got line %d", loc.Range.Start.Line)
	}
}

// TestDefinitionURIMatchesDocument verifies the returned URI is the
// same as the snapshot's URI - we never invent a different URI
// (single-file project today; multi-file is Phase 8's FileID work
// but not yet exercised here).
func TestDefinitionURIMatchesDocument(t *testing.T) {
	loc := definitionFixture(t, "x = 1\nprint(x)\n", lsp.NewPos(1, 6))
	if loc == nil {
		t.Fatal("expected location")
	}
	if loc.Uri != "file:///def_test.rad" {
		t.Errorf("uri: got %q, want %q", loc.Uri, "file:///def_test.rad")
	}
}

// TestDefinitionNoSnapshotReturnsNil verifies the nil-snapshot path.
func TestDefinitionNoSnapshotReturnsNil(t *testing.T) {
	s := NewState()
	loc, err := s.Definition(nil, lsp.NewPos(0, 0))
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	if loc != nil {
		t.Errorf("expected nil location for nil snapshot, got %+v", loc)
	}
}
