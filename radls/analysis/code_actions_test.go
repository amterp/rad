package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// Behavior cases (shebang insertion, null-union quickfix,
// range-scoping) live in radls/lstesting/snapshots/code_action.snap.
// This file keeps the pure-function unit tests - they verify the
// helpers used by structuredFixFor and rangesOverlap, where a
// snapshot would just be one apply-and-check per case without
// adding wire-level confidence.

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
