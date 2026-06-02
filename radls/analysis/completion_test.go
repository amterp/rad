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

// TestCompletionUFCSRanksRelevantBuiltinsAboveMidEdit verifies the
// last-good fallback for UFCS ranking. The natural typing flow is:
// user has a clean script, then types `xs.` and IMMEDIATELY hits
// completion - the converter bails on the trailing dot, current
// snapshot's resolved/types are nil. Without the fallback, builtin
// ranking degraded to plain alphabetical at the very moment ranking
// would be most useful. With the fallback, we look up `xs`'s type
// against the last-good snapshot's indexes and the int-list-shaped
// builtins still float to the relevant tier.
func TestCompletionUFCSRanksRelevantBuiltinsAboveMidEdit(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///comp_midedit_test.rad"
	// Open a clean script so we have a last-good version with
	// resolved/types populated.
	s.AddDoc(uri, "xs: int[] = [1, 2, 3]\n")
	// Simulate the user typing `xs.` - mid-edit, the trailing dot
	// makes the CST->AST converter bail and the new snapshot's
	// resolved/types are nil.
	s.UpdateDoc(uri, []lsp.TextDocumentContentChangeEvent{
		{Text: "xs: int[] = [1, 2, 3]\nxs.\n"},
	})
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	defer snap.Release()
	// Sanity check: this version really is mid-edit (no resolved).
	// If this ever flips to non-nil, the test is no longer
	// exercising the fallback path - rewrite the fixture.
	if snap.resolved != nil && snap.types != nil {
		t.Skip("snapshot converted cleanly - mid-edit fallback path " +
			"not exercised by this fixture; rewrite if grammar recovers " +
			"trailing-dot now")
	}
	// And the last-good is the prior version with usable indexes.
	if !hasUsableResolved(snap.LastGood()) {
		t.Fatal("expected last-good to carry resolved/types after a " +
			"prior good version was registered")
	}

	items, err := s.Complete(snap, lsp.NewPos(1, 3))
	if err != nil {
		t.Fatalf("Complete: %v", err)
	}

	want := map[string]bool{"len": true, "sort": true}
	for _, it := range items {
		if !want[it.Label] {
			continue
		}
		if it.SortText != sortTierBuiltinRelevant {
			t.Errorf("%q mid-edit: SortText=%q, want relevant tier %q",
				it.Label, it.SortText, sortTierBuiltinRelevant)
		}
		delete(want, it.Label)
	}
	for label := range want {
		t.Errorf("%q missing from completion list", label)
	}
}

// TestCompletionUFCSMidEditNoPriorGoodDegradesGracefully verifies
// that opening a document that's broken on the first version (no
// prior version was ever good) does NOT crash and just skips UFCS
// ranking - the fallback is opportunistic, not required.
func TestCompletionUFCSMidEditNoPriorGoodDegradesGracefully(t *testing.T) {
	// Open directly into a mid-edit state with no prior good version.
	items := completionFixture(t, "xs: int[] = [1, 2, 3]\nxs.\n", lsp.NewPos(1, 3))
	// Builtins should still appear; we just don't assert promotion
	// to the relevant tier (no last-good to consult).
	gotLen := false
	for _, it := range items {
		if it.Label == "len" {
			gotLen = true
			break
		}
	}
	if !gotLen {
		t.Fatal("expected 'len' to appear in completion list")
	}
}

// TestCompletionCallsOffersFunctionsOnly verifies the function-
// reference slot: at `calls <prefix>` only functions are valid, so
// the popup offers top-level fns, function-valued top-level vars, and
// builtin functions - and excludes plain vars and command args, which
// the flat completion list would otherwise surface.
func TestCompletionCallsOffersFunctionsOnly(t *testing.T) {
	// Command (with the callback being typed) comes first; the
	// callback targets are defined below. That's the structure real
	// scripts use, and a callback sees every top-level binding since
	// it runs after the whole top-level executes - so do_deploy,
	// handler, and flag's sibling are all in scope here.
	src := "command run:\n" +
		"    env str\n" +
		"    calls do_de\n" +
		"\n" +
		"fn do_deploy():\n" +
		"    print(\"hi\")\n" +
		"\n" +
		"handler = fn():\n" +
		"    print(\"h\")\n" +
		"\n" +
		"flag = \"x\"\n"
	// Cursor sits right after `do_de` on the `calls` line (line 2).
	items := completionFixture(t, src, lsp.NewPos(2, 15))

	labels := make(map[string]lsp.CompletionItem)
	for _, it := range items {
		labels[it.Label] = it
	}

	// Functions are offered: a top-level fn, a function-valued var, and
	// a builtin - all valid callbacks.
	for _, want := range []string{"do_deploy", "handler", "print"} {
		it, ok := labels[want]
		if !ok {
			t.Errorf("expected %q to be offered as a function callback", want)
			continue
		}
		if it.Kind != lsp.CompletionKindFunction {
			t.Errorf("%q: Kind=%d, want Function(%d)", want, it.Kind, lsp.CompletionKindFunction)
		}
	}

	// Non-function names are NOT offered: a plain string var and a
	// command arg can't be invoked as a callback.
	for _, unwanted := range []string{"flag", "env"} {
		if _, ok := labels[unwanted]; ok {
			t.Errorf("%q should not be offered in a `calls` callback slot", unwanted)
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
