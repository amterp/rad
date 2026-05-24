package analysis

import (
	"strings"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

func codeActionFixture(t *testing.T, src string, r lsp.Range) []lsp.CodeAction {
	t.Helper()
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///ca_test.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	defer snap.Release()
	actions, err := s.CodeAction(snap, r)
	if err != nil {
		t.Fatalf("CodeAction: %v", err)
	}
	return actions
}

// TestQuickFixForNullInUnion verifies the structured fix lights up
// for `T|null` and produces a workspace edit. The typed-assign
// form (`x: int|null = ...`) is the path where the binder's
// null-union heuristic fires today - the fn-param form goes
// through a different parent and doesn't (yet) produce the
// suggestion. We test what we have; the fn-param form is a
// natural follow-up once the binder grows its heuristic.
func TestQuickFixForNullInUnion(t *testing.T) {
	src := "x: int|null = 5\n"
	// Whole-line range so the overlap check matches any column.
	r := lsp.NewLineRange(0, 0, 20)
	actions := codeActionFixture(t, src, r)

	found := false
	for _, a := range actions {
		if a.Kind == lsp.CodeActionQuickFix && a.Edit != nil {
			for _, edits := range a.Edit.Changes {
				for _, e := range edits {
					if e.NewText == "?" {
						found = true
					}
				}
			}
		}
	}
	if !found {
		t.Errorf("expected quickfix that replaces span with '?', got %d actions: %+v",
			len(actions), actions)
	}
}

// TestShebangActionAlwaysOffered verifies the shebang insertion
// action shows up regardless of what other quick fixes are
// present. It's the "fresh-file" guidance and shouldn't drop off
// the menu just because the file has diagnostics.
func TestShebangActionAlwaysOffered(t *testing.T) {
	src := "fn f(x: int|null) -> int: return 0\n"
	actions := codeActionFixture(t, src, lsp.NewLineRange(0, 0, 1))
	foundShebang := false
	for _, a := range actions {
		if strings.Contains(strings.ToLower(a.Title), "shebang") {
			foundShebang = true
		}
	}
	if !foundShebang {
		t.Errorf("expected shebang action, got %d actions: %+v",
			len(actions), actions)
	}
}

// TestCodeActionsScopedByRange verifies diagnostics outside the
// request range don't surface their quick fixes. Editors send the
// selection range and expect only-relevant actions; returning
// every diagnostic's fix on every request would clutter the menu.
func TestCodeActionsScopedByRange(t *testing.T) {
	// Two type errors on different lines. The first is on line 0;
	// asking for actions in a range on line 5 should NOT include
	// it.
	src := "fn f(x: int|null) -> int: return 0\n\n\n\n\n\nx = 1\n"
	// Ask for actions only on the last line (no diagnostic there).
	actions := codeActionFixture(t, src, lsp.NewLineRange(6, 0, 5))

	for _, a := range actions {
		if a.Kind == lsp.CodeActionQuickFix && a.Edit != nil {
			t.Errorf("unexpected quickfix at line 6: %+v", a)
		}
	}
}

// TestBuildNullUnionFix verifies the helper handles the
// trailing-null shapes the ERROR span actually contains today.
// Other shapes (leading null, multi-type unions, bare null)
// don't have safe in-place rewrites - we'd need to widen the
// span - so they return false. Those land as follow-ups.
func TestBuildNullUnionFix(t *testing.T) {
	cases := []struct {
		in   string
		out  string
		want bool
	}{
		{"|null", "?", true},
		{"| null", "?", true},
		{"|null|", "?", true},   // trailing pipe from parser
		{"null", "", false},     // bare null - no safe in-place fix
		{"null|int", "", false}, // leading null - needs widening
		{"int", "", false},      // no null at all
	}
	for _, c := range cases {
		got, ok := buildNullUnionFix(c.in)
		if ok != c.want {
			t.Errorf("buildNullUnionFix(%q) ok=%v, want %v", c.in, ok, c.want)
		}
		if ok && got != c.out {
			t.Errorf("buildNullUnionFix(%q) got %q, want %q", c.in, got, c.out)
		}
	}
}

// TestRangesOverlap checks the helper's edge cases - touching at
// a boundary doesn't count as overlap (LSP ranges are half-open
// at the end), strict before / after both return false.
func TestRangesOverlap(t *testing.T) {
	cases := []struct {
		name string
		a, b lsp.Range
		want bool
	}{
		{"identical", lsp.NewLineRange(0, 0, 5), lsp.NewLineRange(0, 0, 5), true},
		{"a-inside-b", lsp.NewLineRange(0, 2, 3), lsp.NewLineRange(0, 0, 10), true},
		{"touch-at-edge", lsp.NewLineRange(0, 0, 5), lsp.NewLineRange(0, 5, 10), false},
		{"disjoint", lsp.NewLineRange(0, 0, 3), lsp.NewLineRange(0, 10, 20), false},
		{"diff-lines", lsp.NewLineRange(0, 0, 5), lsp.NewLineRange(2, 0, 5), false},
	}
	for _, c := range cases {
		if got := rangesOverlap(c.a, c.b); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}
