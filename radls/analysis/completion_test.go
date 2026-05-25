package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// Most completion behavior cases (shebang, args block decls,
// fn-body params + locals, position filtering) live in
// radls/lstesting/snapshots/completion.snap. This file keeps the
// invariant assertions that don't read well as snapshots:
//
//  - SortText is monotonic across the (huge) returned list.
//  - No duplicate Labels.
//  - Scope tiers ("0"=enclosing, "1"=file, "2"=builtins) attach
//    to the right items.
//
// Reading a 600-line snapshot to verify "monotonic SortText" is
// impractical; the Go assertion makes the rule explicit.

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
	items := completionFixture(t, "x = (", lsp.NewPos(0, 5))
	prev := ""
	for _, it := range items {
		if prev != "" && it.SortText < prev {
			t.Errorf("nil-AST path not sorted by SortText: %q before %q",
				prev, it.SortText)
		}
		prev = it.SortText
	}
}

// TestCompletionUFCSRanksRelevantBuiltinsAbove verifies that at
// `xs.<cursor>`, builtins whose first param accepts the receiver
// type (e.g. `len` accepts a list) sort above unrelated builtins.
// The win: typing a `.` after a typed local floats the relevant
// completions to the top instead of leaving them buried in the
// alphabetical builtin tier.
func TestCompletionUFCSRanksRelevantBuiltinsAbove(t *testing.T) {
	// `xs.len()` is a complete UFCS call. Completion at the cursor
	// position right after the `.` is the natural place to test
	// ranking: the receiver is `xs` (typed `int[]`), and builtins
	// whose first param accepts a list should outrank unrelated
	// ones.
	src := "xs: int[] = [1, 2, 3]\nxs.len()\n"
	// Cursor sits right after the `.` on line 1, col 3.
	items := completionFixture(t, src, lsp.NewPos(1, 3))

	want := map[string]bool{"len": true, "sort": true}
	gotRelevant := make(map[string]string)
	for _, it := range items {
		if !want[it.Label] {
			continue
		}
		gotRelevant[it.Label] = it.SortText
	}
	for label := range want {
		got, ok := gotRelevant[label]
		if !ok {
			t.Errorf("%q missing from completion list", label)
			continue
		}
		if got != sortTierBuiltinRelevant {
			t.Errorf("%q: SortText=%q, want relevant tier %q",
				label, got, sortTierBuiltinRelevant)
		}
	}

	// Unrelated builtins (e.g. `now`, `parse_int`) should stay at
	// the plain builtin tier - the receiver type isn't compatible
	// with their first param.
	unrelated := map[string]bool{"now": true, "parse_int": true}
	for _, it := range items {
		if !unrelated[it.Label] {
			continue
		}
		if it.SortText == sortTierBuiltinRelevant {
			t.Errorf("%q unexpectedly promoted to relevant tier", it.Label)
		}
	}
}

// TestCompletionEmptySnapshotReturnsNil verifies nil-snapshot
// path. Unreachable through the wire harness; defensive.
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
