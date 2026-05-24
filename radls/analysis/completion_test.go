package analysis

import (
	"strings"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

func completionFixture(t *testing.T, src string, pos lsp.Pos) []lsp.CompletionItem {
	t.Helper()
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///comp_test.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	defer snap.Release()
	items, err := s.Complete(snap, pos)
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	return items
}

// hasLabel reports whether `items` contains a CompletionItem with
// the given label. Useful for "X should be present" assertions
// without coupling to the full alphabetical order.
func hasLabel(items []lsp.CompletionItem, label string) bool {
	for _, it := range items {
		if it.Label == label {
			return true
		}
	}
	return false
}

// TestCompletionIncludesBuiltins verifies builtins like `print`
// show up as Function-kind completions, with a non-empty Detail
// (the signature). This is the bedrock of "the user can discover
// API surface from completion."
func TestCompletionIncludesBuiltins(t *testing.T) {
	items := completionFixture(t, "x = ", lsp.NewPos(0, 4))
	if !hasLabel(items, "print") {
		t.Errorf("expected 'print' builtin in completions")
	}
	for _, it := range items {
		if it.Label == "print" {
			if it.Kind != lsp.CompletionKindFunction {
				t.Errorf("print: kind=%d, want Function", it.Kind)
			}
			if !strings.Contains(it.Detail, "->") {
				t.Errorf("print detail missing signature arrow: %q",
					it.Detail)
			}
		}
	}
}

// TestCompletionIncludesTopLevelDecls verifies top-level vars and
// fns from the file appear as Variable / Function completions.
func TestCompletionIncludesTopLevelDecls(t *testing.T) {
	src := "alpha = 1\n\nfn beta():\n    print(\"hi\")\n\n"
	items := completionFixture(t, src, lsp.NewPos(4, 0))
	if !hasLabel(items, "alpha") {
		t.Errorf("expected 'alpha' top-level var in completions")
	}
	if !hasLabel(items, "beta") {
		t.Errorf("expected 'beta' top-level fn in completions")
	}
}

// TestCompletionIncludesArgsBlockNames verifies args: block decls
// surface in completions - they're file-scope ambient bindings the
// runtime populates from CLI flags, and users reach for them
// constantly.
func TestCompletionIncludesArgsBlockNames(t *testing.T) {
	src := "args:\n    name str\n    age int = 30\n\n"
	items := completionFixture(t, src, lsp.NewPos(3, 0))
	if !hasLabel(items, "name") {
		t.Errorf("expected 'name' arg in completions")
	}
	if !hasLabel(items, "age") {
		t.Errorf("expected 'age' arg in completions")
	}
}

// TestCompletionIncludesEnclosingFnParams verifies the params of
// the enclosing function appear when the cursor is inside that
// function's body.
func TestCompletionIncludesEnclosingFnParams(t *testing.T) {
	src := "fn greet(who: str):\n    print(who)\n"
	// Cursor on line 1 col 10 (inside print's arg).
	items := completionFixture(t, src, lsp.NewPos(1, 10))
	if !hasLabel(items, "who") {
		t.Errorf("expected 'who' param in completions inside greet body")
	}
}

// TestCompletionExcludesLocalsDeclaredAfterCursor verifies a local
// declared LATER in the body isn't suggested - the user can't
// reference it yet at this cursor position.
func TestCompletionExcludesLocalsDeclaredAfterCursor(t *testing.T) {
	src := "fn f():\n    x = 1\n    y = 2\n"
	// Cursor at line 1 col 8 (right after `x = 1`). y is declared
	// on line 2.
	items := completionFixture(t, src, lsp.NewPos(1, 8))
	if hasLabel(items, "y") {
		t.Errorf("'y' is declared after cursor, shouldn't be in completions")
	}
}

// TestCompletionDedupedAndSorted verifies labels are unique and
// the popup order is (SortText, Label). Without the SortText
// tier, alphabetical-only sort would bury locals under builtins
// that happen to share a prefix.
func TestCompletionDedupedAndSorted(t *testing.T) {
	items := completionFixture(t, "x = 1\n", lsp.NewPos(1, 0))
	seen := make(map[string]bool)
	type key struct{ sort, label string }
	var prev key
	first := true
	for _, it := range items {
		if seen[it.Label] {
			t.Errorf("duplicate label: %q", it.Label)
		}
		seen[it.Label] = true
		cur := key{it.SortText, it.Label}
		if !first {
			if cur.sort < prev.sort {
				t.Errorf("SortText regressed: %q before %q", prev.sort, cur.sort)
			}
			if cur.sort == prev.sort && cur.label < prev.label {
				t.Errorf("Label not sorted within tier %q: %q before %q",
					cur.sort, prev.label, cur.label)
			}
		}
		first = false
		prev = cur
	}
}

// TestCompletionScopeTiers verifies the tier assignment: locals
// get "0", file-scope gets "1", builtins get "2". This is the
// load-bearing UX win - locals at the top, builtins at the
// bottom, the editor's filter doesn't bury what the user just
// typed.
func TestCompletionScopeTiers(t *testing.T) {
	src := "alpha = 1\n\nfn beta(who: str):\n    local = 2\n    print(local)\n"
	// Cursor inside fn beta at line 4 col 4.
	items := completionFixture(t, src, lsp.NewPos(4, 4))

	wantTier := map[string]string{
		"who":   "0", // enclosing-fn param
		"local": "0", // earlier-local
		"alpha": "1", // file-scope var
		"beta":  "1", // file-scope fn
		"print": "2", // builtin
	}
	for _, it := range items {
		want, tracked := wantTier[it.Label]
		if !tracked {
			continue
		}
		if it.SortText != want {
			t.Errorf("%q: SortText=%q, want %q", it.Label, it.SortText, want)
		}
	}
}

// TestCompletionNilASTSorted verifies the parse-failed path still
// returns a sorted list. Before the fix, hitting an ERROR-node
// state mid-edit (the most common typing state) skipped the
// sort entirely and the popup reordered itself randomly per
// keystroke.
func TestCompletionNilASTSorted(t *testing.T) {
	// Unclosed expression - the converter typically returns nil AST.
	items := completionFixture(t, "x = (", lsp.NewPos(0, 5))
	// We don't strictly require nil AST here; we require sorted
	// output regardless of the AST path the snapshot took.
	prev := ""
	for _, it := range items {
		if prev != "" && it.SortText < prev {
			t.Errorf("nil-AST path not sorted by SortText: %q before %q",
				prev, it.SortText)
		}
		prev = it.SortText
	}
}

// TestCompletionShebangFirstOnLineZero verifies the shebang stub
// stays at the very front of the list on line 0 - that's the
// "new empty file" experience we don't want to lose.
func TestCompletionShebangFirstOnLineZero(t *testing.T) {
	items := completionFixture(t, "", lsp.NewPos(0, 0))
	if len(items) == 0 {
		t.Fatal("expected completions")
	}
	if !strings.HasPrefix(items[0].Label, "#!") {
		t.Errorf("expected shebang first, got %q", items[0].Label)
	}
}

// TestCompletionEmptySnapshotReturnsNil verifies nil snapshot
// path. Server's nil-check should prevent this, but defensive is
// cheap.
func TestCompletionEmptySnapshotReturnsNil(t *testing.T) {
	s := NewState()
	items, err := s.Complete(nil, lsp.NewPos(0, 0))
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}
	if items != nil {
		t.Errorf("expected nil for nil snapshot, got %d items", len(items))
	}
}
