package radfmt

import (
	"strings"

	"github.com/amterp/rad/rts"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// structuralSig builds a fingerprint of a tree's *code* structure: the kinds and
// field names of every named node, plus every anonymous token that carries a
// field name, in pre-order. It deliberately ignores *fieldless* anonymous
// punctuation tokens (so a canonical trailing comma is allowed), string quote
// characters (a string node is a string node regardless of quote), whitespace,
// and positions. Comments are excluded here and counted separately, since
// formatting may legitimately move a comment but must never drop one or change
// what the code parses to.
//
// A field-bearing anonymous token is semantically load-bearing - the for-loop
// `with <ctx>` identifier, `?`/`*` arg markers, keyword choices like
// defer/errdefer or quiet/confirm, and operators - so its presence and kind are
// recorded (`field=kind`). This is what stops a construct formatter from
// silently dropping such a clause: doing so changes the signature and trips the
// equivalence guard. Recording the kind (the literal, for keyword tokens) also
// catches a swap where the choice rides on the token rather than the node kind:
// `defer` and `errdefer` are one `defer_block` kind distinguished only by their
// `keyword` token, so without this the named-node path alone can't tell them apart.
func structuralSig(n *ts.Node) (sig string, comments int) {
	var sb strings.Builder
	var walk func(field string, n *ts.Node)
	walk = func(field string, n *ts.Node) {
		if isComment(n) {
			comments++
			return
		}
		// Record named nodes and field-bearing anonymous tokens alike, each
		// opening a `(...)` group so its children nest under it (a flat encoding
		// would let a child of an anonymous token masquerade as its sibling).
		// Fieldless anonymous punctuation is skipped but still recursed through,
		// keeping trailing-comma-style canonicalization invisible to the guard.
		record := n.IsNamed() || field != ""
		if record {
			if field != "" {
				sb.WriteString(field)
				sb.WriteByte('=')
			}
			sb.WriteString(n.Kind())
			sb.WriteByte('(')
		}
		cursor := n.Walk()
		defer cursor.Close()
		for i, ch := range n.Children(cursor) {
			walk(n.FieldNameForChild(uint32(i)), &ch)
		}
		if record {
			sb.WriteByte(')')
		}
	}
	walk("", n)
	return sb.String(), comments
}

// structuralDump renders a readable, position-free tree of named node kinds and
// their field names, plus field-bearing anonymous tokens (comments shown as a
// normalized marker). It's the human-friendly twin of structuralSig - it records
// exactly what the signature does - so tests diff the dump of the input against
// the dump of the formatted output to prove formatting never changed the code's
// structure (including dropped optional clauses), with a clear side-by-side
// failure.
func structuralDump(n *ts.Node) string {
	var sb strings.Builder
	var walk func(depth int, field string, n *ts.Node)
	walk = func(depth int, field string, n *ts.Node) {
		if isComment(n) {
			sb.WriteString(strings.Repeat("  ", depth))
			sb.WriteString("<comment>\n")
			return
		}
		// Mirror structuralSig: named nodes and field-bearing anonymous tokens
		// both print a line and indent their children; fieldless punctuation is
		// skipped. Anonymous tokens get a `(token)` marker so the dump reads
		// clearly.
		if n.IsNamed() || field != "" {
			sb.WriteString(strings.Repeat("  ", depth))
			if field != "" {
				sb.WriteString(field)
				sb.WriteString(": ")
			}
			sb.WriteString(n.Kind())
			if !n.IsNamed() {
				sb.WriteString(" (token)")
			}
			sb.WriteByte('\n')
			depth++
		}
		cursor := n.Walk()
		defer cursor.Close()
		for i, ch := range n.Children(cursor) {
			walk(depth, n.FieldNameForChild(uint32(i)), &ch)
		}
	}
	walk(0, "", n)
	return sb.String()
}

// structurallyEquivalent reports whether formatted parses cleanly to the same
// code structure (and same comment count) as the original, captured in wantSig /
// wantComments. This is Format's last line of defense: if a construct formatter
// has a bug that would change what the code means or drop a comment, Format
// detects it here and falls back to returning the original untouched.
func structurallyEquivalent(formatted, wantSig string, wantComments int) bool {
	parser, err := rts.NewRadParser()
	if err != nil {
		return false
	}
	defer parser.Close()

	tree := parser.Parse(formatted)
	if tree.HasInvalidNodes() {
		return false
	}
	root := tree.Root()
	if root == nil {
		return false
	}
	gotSig, gotComments := structuralSig(root)
	return gotSig == wantSig && gotComments == wantComments
}
