package radfmt

import (
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// children returns all child nodes (named, anonymous tokens, and extras such as
// comments) in source order. Comment extras appear here, which is how the
// statement sequencer and construct formatters see and place them.
func children(n *ts.Node) []ts.Node {
	cursor := n.Walk()
	defer cursor.Close()
	return n.Children(cursor)
}

// childByField returns the single child stored under fieldName, or nil.
func childByField(n *ts.Node, fieldName string) *ts.Node {
	return n.ChildByFieldName(fieldName)
}

// isComment reports whether n is a `//` comment node.
func isComment(n *ts.Node) bool { return n.Kind() == rl.K_COMMENT }

// nodeText returns the exact source span of n.
func (p *printer) nodeText(n *ts.Node) string {
	return p.src[n.StartByte():n.EndByte()]
}

// startRow / endRow are the 0-based source rows a node spans.
func startRow(n *ts.Node) uint { return n.StartPosition().Row }
func endRow(n *ts.Node) uint   { return n.EndPosition().Row }

// blankBetween reports whether at least one fully blank source line separates a
// (above) from b (below).
func blankBetween(a, b *ts.Node) bool {
	return startRow(b) > endRow(a)+1
}

// sameRow reports whether b begins on the row a ends on (used to detect trailing
// same-line comments).
func sameRow(a, b *ts.Node) bool {
	return startRow(b) == endRow(a)
}

// containsComment reports whether n's subtree contains any comment node. Used as
// a safety guard: a construct formatter that doesn't explicitly handle interior
// comments falls back to verbatim when one is present, so comments are never
// dropped.
func containsComment(n *ts.Node) bool {
	if isComment(n) {
		return true
	}
	for _, ch := range children(n) {
		if containsComment(&ch) {
			return true
		}
	}
	return false
}
