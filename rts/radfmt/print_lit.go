package radfmt

import (
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// formatList renders `[a, b, c]`, wrapping one-element-per-line with a trailing
// comma when it exceeds the line width.
func (p *printer) formatList(n *ts.Node) Doc {
	var items []Doc
	for _, c := range childPtrs(n) {
		switch c.Kind() {
		case tLBracket, tRBracket, tComma:
			continue
		}
		if c.IsNamed() {
			items = append(items, p.formatExpr(c))
		}
	}
	return p.delimited(tLBracket, tRBracket, items)
}

// formatMap renders `{key: value, ...}`, wrapping one-entry-per-line when it
// exceeds the line width. Keys keep their original form (string or bareword);
// the canonical shape is `key: value` with a single space after the colon.
func (p *printer) formatMap(n *ts.Node) Doc {
	var entries []Doc
	for _, c := range childPtrs(n) {
		if c.Kind() == kMapEntry {
			entries = append(entries, p.formatMapEntry(c))
		}
	}
	return p.delimited("{", "}", entries)
}

func (p *printer) formatMapEntry(n *ts.Node) Doc {
	key := childByField(n, rl.F_KEY)
	value := childByField(n, rl.F_VALUE)
	if key == nil || value == nil {
		return p.verbatim(n)
	}
	return concat(p.formatExpr(key), text(": "), p.formatExpr(value))
}
