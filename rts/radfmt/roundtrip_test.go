package radfmt

import (
	"testing"

	gd "github.com/amterp/go-delta"
)

// TestRoundTripOptionalClauses is a property test for the class of bug behind
// issue #144: a formatter silently dropping an optional clause that the grammar
// realizes as a bare (anonymous) token. For a spread of productions that carry
// such tokens, it asserts the parse -> fmt -> parse round-trip preserves the
// code structure - now including field-bearing anonymous tokens, since that is
// the dimension the old structural guard was blind to (see safety.go).
//
// The for-loop `with <ctx>` clause is the live case: drop it and the dump loses
// `context: identifier (token)`, failing here.
//
// Two kinds of case below. The "exercised" group feeds deliberately messy input
// through an implemented formatter, so the formatter actually rewrites the source
// and the round-trip proves the rewrite preserved structure (the real test). The
// "verbatim lock" group covers constructs with no formatter yet (defer):
// formatRaw returns them byte-for-byte, so they pass trivially today - they exist
// to (a) assert the strengthened guard doesn't false-positive on these
// field-bearing tokens and (b) fail loudly the day one graduates to a real
// formatter that drops a clause. When that day comes, give the case messy input
// so it joins the exercised group. Inputs are valid Rad drawn from the syntax
// snapshot corpus (rts/test/st_snapshots).
func TestRoundTripOptionalClauses(t *testing.T) {
	cases := []struct {
		name string
		src  string
	}{
		// --- Exercised: messy input an implemented formatter must rewrite. ---
		// For-loop context clause - the #144 bug and its keyed variant.
		{"for_with_context", "for item in items  with  ctx:\n    print(ctx.idx, item)\n"},
		{"keyed_for_with_context", "for k,v in m with ctx:\n    print(ctx.idx, k, v)\n"},
		{"for_no_context", "for  item  in  items:\n    print(item)\n"},
		// Operators and markers are all anonymous field-bearing tokens.
		{"ternary", "y = a?b:c\n"},
		{"binary_ops", "z = 1+2*3 and a or b\n"},
		{"unary_ops", "z = not  a\nw = -  b\n"},
		{"compound_and_incr", "x+=1\ni++\n"},
		{"typed_assign", "x:int=1\n"},
		// list_comprehension is verbatim, but the surrounding assign is formatted,
		// so messy `=` spacing makes this a real round-trip over a verbatim subtree.
		{"list_comprehension", "area=[width[i] * height[i] for i in range(width)]\n"},
		// Optional `?` marker on an arg declaration - the args formatter tightens
		// the spacing, so the round-trip proves the marker survives that rewrite.
		{"args_optional_marker", "args:\n    count   int ?\n"},

		// --- Verbatim locks: no formatter yet, so these are no-ops today. ---
		// defer vs errdefer share the `defer_block` kind, distinguished only by the
		// bare keyword token the strengthened guard now records.
		{"defer_keyword", "defer print(1)\n"},
		{"errdefer_keyword", "errdefer print(2)\n"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			src := normalizeLineEndings(tc.src)
			raw, _, _, ok := formatRaw(src)
			if !ok {
				t.Fatalf("input failed to parse cleanly: %s", tc.name)
			}
			before := dumpStructure(t, src)
			after := dumpStructure(t, raw)
			if before != after {
				t.Errorf("parse -> fmt -> parse changed structure for %s:\n%s",
					tc.name,
					gd.DiffWith(before, after,
						gd.WithColor(true),
						gd.WithLayout(gd.LayoutPreferSideBySide),
						gd.WithWidth(120)))
			}
		})
	}
}
