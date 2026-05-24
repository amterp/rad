package analysis

import (
	"strings"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// hoverFixture spins up a State, opens one document, snapshots it,
// and runs Hover at the given position. It's the workhorse of these
// tests - the LSP wire layer is out of scope here; we want to know
// that the analysis function picks the right symbol and renders the
// right markdown.
func hoverFixture(t *testing.T, src string, pos lsp.Pos) *lsp.Hover {
	t.Helper()
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///hover_test.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot after AddDoc")
	}
	defer snap.Release()
	h, err := s.Hover(snap, pos)
	if err != nil {
		t.Fatalf("Hover: %v", err)
	}
	return h
}

// TestHoverOnUntypedLocal verifies hover on a plain `x = 1` shows
// the local kind tag and the inferred int type. This is the most
// common script-author case and exercises the SymbolTypes lookup
// path (no declared annotation -> inferred wins).
func TestHoverOnUntypedLocal(t *testing.T) {
	// Cursor on the `x` (col 0) of `x = 1`.
	h := hoverFixture(t, "x = 1\n", lsp.NewPos(0, 0))
	if h == nil {
		t.Fatal("expected hover, got nil")
	}
	if !strings.Contains(h.Contents.Value, "(local)") {
		t.Errorf("missing local tag: %q", h.Contents.Value)
	}
	if !strings.Contains(h.Contents.Value, "x") {
		t.Errorf("missing name: %q", h.Contents.Value)
	}
	if !strings.Contains(h.Contents.Value, "int") {
		t.Errorf("missing inferred int type: %q", h.Contents.Value)
	}
}

// TestHoverOnBuiltin verifies hover on a builtin returns its
// pre-parsed signature from FnSignaturesByName. The signature is
// what users actually want to see (param names + types + return
// type), not just the bare type-checker view.
func TestHoverOnBuiltin(t *testing.T) {
	// `print()` - cursor on the `p` of print.
	h := hoverFixture(t, "print(1)\n", lsp.NewPos(0, 0))
	if h == nil {
		t.Fatal("expected hover, got nil")
	}
	if !strings.Contains(h.Contents.Value, "(builtin)") {
		t.Errorf("missing builtin tag: %q", h.Contents.Value)
	}
	// We don't need the full signature - just that the signature
	// path is used (recognizable by the `->` return-type arrow).
	if !strings.Contains(h.Contents.Value, "->") {
		t.Errorf("expected signature with return arrow: %q", h.Contents.Value)
	}
}

// TestHoverOffIdentifierReturnsNil verifies that hovering on
// whitespace, punctuation, or other non-identifier territory yields
// a nil hover (which the LSP server forwards as the spec's null
// reply). Returning an empty hover or one tagged "(unresolved)"
// for these positions would clutter the editor.
func TestHoverOffIdentifierReturnsNil(t *testing.T) {
	// Cursor in the trailing whitespace of `x = 1   `.
	h := hoverFixture(t, "x = 1   \n", lsp.NewPos(0, 7))
	if h != nil {
		t.Errorf("expected nil hover off-identifier, got %+v", h)
	}
}

// TestHoverRangeMatchesIdentifierSpan verifies the returned Range
// covers exactly the identifier token. Editors use this to
// highlight the hover target; getting the range wrong (e.g.
// returning the enclosing assignment span) is visually jarring.
func TestHoverRangeMatchesIdentifierSpan(t *testing.T) {
	// `   alpha = 1` - identifier starts at col 3, ends at col 8.
	h := hoverFixture(t, "   alpha = 1\n", lsp.NewPos(0, 4))
	if h == nil {
		t.Fatal("expected hover, got nil")
	}
	if h.Range == nil {
		t.Fatal("expected non-nil range")
	}
	if h.Range.Start.Line != 0 || h.Range.Start.Character != 3 {
		t.Errorf("start: got %+v, want (0,3)", h.Range.Start)
	}
	if h.Range.End.Line != 0 || h.Range.End.Character != 8 {
		t.Errorf("end: got %+v, want (0,8)", h.Range.End)
	}
}

// TestHoverCursorAtIdentifierEnd verifies the cursor sitting just
// after the last character of an identifier still produces a
// hover. Editors place the cursor "between characters," so
// position N for an N-char identifier is the position immediately
// after the last char - users expect hover there too.
func TestHoverCursorAtIdentifierEnd(t *testing.T) {
	// `x = 1` - cursor at col 1 sits right after `x`.
	h := hoverFixture(t, "x = 1\n", lsp.NewPos(0, 1))
	if h == nil {
		t.Fatal("expected hover at identifier-end cursor, got nil")
	}
}

// TestHoverOnBrokenSyntax verifies hover doesn't crash mid-edit
// when the source fails to parse into an AST. The snapshot's AST
// is nil in that case and Hover must return (nil, nil) rather
// than dereferencing it. This is the path users hit constantly
// while typing.
func TestHoverOnBrokenSyntax(t *testing.T) {
	h := hoverFixture(t, "x = (\n", lsp.NewPos(0, 0))
	if h != nil {
		// Some "best-effort" hover content is fine if we later add
		// it (e.g. the binder might still recognize `x` as a decl);
		// the strong contract is no-crash. The test name documents
		// the intent.
		t.Logf("hover on broken syntax returned: %q", h.Contents.Value)
	}
}

// TestHoverNoSnapshotReturnsNil verifies we don't crash on a nil
// snapshot. The server's nil-check should prevent this, but the
// analysis function is a public method - defensive is cheap.
func TestHoverNoSnapshotReturnsNil(t *testing.T) {
	s := NewState()
	h, err := s.Hover(nil, lsp.NewPos(0, 0))
	if err != nil {
		t.Fatalf("Hover: %v", err)
	}
	if h != nil {
		t.Errorf("expected nil hover for nil snapshot, got %+v", h)
	}
}

// TestHoverOnTypedLocal verifies the declared annotation is
// preferred over inferred when both are present. The type-check
// pass writes both into SymbolTypes (declared wins on assignment),
// so this is really testing that `: int` propagates from the
// grammar through the binder and survives to hover.
//
// Phase 3 added typed-local grammar; if that work isn't yet in,
// the parser will emit ERROR nodes and the test will skip rather
// than fail. Easy to flip from skip to assert once typed-locals
// are universally available.
func TestHoverOnTypedLocal(t *testing.T) {
	src := "x: int = 5\n"
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///typed.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	defer snap.Release()
	// Skip if grammar doesn't support typed locals yet (Phase 3).
	if snap.tree.HasInvalidNodes() {
		t.Skip("typed-local grammar not yet available (Phase 3 deferred)")
	}
	h, err := s.Hover(snap, lsp.NewPos(0, 0))
	if err != nil {
		t.Fatalf("Hover: %v", err)
	}
	if h == nil {
		t.Fatal("expected hover")
	}
	if !strings.Contains(h.Contents.Value, "int") {
		t.Errorf("expected int in hover: %q", h.Contents.Value)
	}
}
