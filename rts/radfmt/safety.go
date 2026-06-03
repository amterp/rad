package radfmt

import (
	"strings"

	"github.com/amterp/rad/rts"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// structuralSig builds a fingerprint of a tree's *code* structure: the kinds and
// field names of every named node, in pre-order. It deliberately ignores
// anonymous punctuation tokens (so a canonical trailing comma is allowed),
// string quote characters (a string node is a string node regardless of quote),
// whitespace, and positions. Comments are excluded here and counted separately,
// since formatting may legitimately move a comment but must never drop one or
// change what the code parses to.
func structuralSig(n *ts.Node) (sig string, comments int) {
	var sb strings.Builder
	var walk func(field string, n *ts.Node)
	walk = func(field string, n *ts.Node) {
		if isComment(n) {
			comments++
			return
		}
		if n.IsNamed() {
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
		if n.IsNamed() {
			sb.WriteByte(')')
		}
	}
	walk("", n)
	return sb.String(), comments
}

// structuralDump renders a readable, position-free tree of named node kinds and
// their field names (comments shown as a normalized marker). It's the
// human-friendly twin of structuralSig: tests diff the dump of the input against
// the dump of the formatted output to prove formatting never changed the code's
// structure, with a clear side-by-side failure.
func structuralDump(n *ts.Node) string {
	var sb strings.Builder
	var walk func(depth int, field string, n *ts.Node)
	walk = func(depth int, field string, n *ts.Node) {
		if isComment(n) {
			sb.WriteString(strings.Repeat("  ", depth))
			sb.WriteString("<comment>\n")
			return
		}
		if n.IsNamed() {
			sb.WriteString(strings.Repeat("  ", depth))
			if field != "" {
				sb.WriteString(field)
				sb.WriteString(": ")
			}
			sb.WriteString(n.Kind())
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
