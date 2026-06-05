package radfmt

import (
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// Node kinds / tokens not exported as rl constants.
const (
	kMapEntry  = "map_entry"
	kEmptyList = "empty_list"
	tColon     = ":"
	tComma     = ","
	tDot       = "."
	tLBracket  = "["
	tRBracket  = "]"
	tLParen    = "("
	tRParen    = ")"
)

// namedChildrenOf returns stable pointers to a node's named children only.
func namedChildrenOf(n *ts.Node) []*ts.Node {
	var out []*ts.Node
	for _, c := range childPtrs(n) {
		if c.IsNamed() {
			out = append(out, c)
		}
	}
	return out
}

// unwrap collapses transparent wrapper nodes: a node whose sole named child
// spans exactly the same bytes carries no syntax of its own (it's a link in the
// expression-precedence chain, or a `literal`/`primary_expr` wrapper). Descend
// until we reach a node that actually contributes structure.
func unwrap(n *ts.Node) *ts.Node {
	for {
		named := namedChildrenOf(n)
		if len(named) == 1 &&
			named[0].StartByte() == n.StartByte() &&
			named[0].EndByte() == n.EndByte() {
			n = named[0]
			continue
		}
		return n
	}
}

// formatExpr formats an expression node, collapsing the precedence-chain
// wrappers first. Anything not explicitly handled - or any subtree containing a
// comment we don't place - falls back to verbatim, which is always safe.
func (p *printer) formatExpr(n *ts.Node) Doc {
	if n == nil {
		return text("")
	}
	n = unwrap(n)

	// Known limitation: comments between statements (and in block bodies) are
	// placed by formatSeq, but comments *inside* an expression aren't attached
	// yet. Rather than risk dropping one, emit any comment-bearing expression
	// verbatim - structurally safe, just not reflowed. Per-construct interior
	// comment handling (see DESIGN.md) can replace this as constructs mature.
	//
	// [F36] an expression containing an interior comment is emitted verbatim
	if containsComment(n) {
		return p.verbatim(n)
	}

	switch n.Kind() {
	// [F33] number/bool/null/identifier literals are preserved verbatim
	case rl.K_IDENTIFIER, rl.K_INT, rl.K_FLOAT, rl.K_BOOL, rl.K_NULL:
		return text(p.nodeText(n))

	case kEmptyList:
		return text("[]")

	case rl.K_STRING:
		// Strings are emitted verbatim for now: contents, escapes, and
		// interpolation expressions are preserved exactly. (Quote-style
		// normalization is a deliberate follow-up.)
		return p.verbatim(n)

	case rl.K_OR_EXPR, rl.K_AND_EXPR, rl.K_COMPARE_EXPR, rl.K_ADD_EXPR, rl.K_MULT_EXPR:
		return p.formatBinary(n)

	case rl.K_UNARY_EXPR:
		return p.formatUnary(n)

	case rl.K_TERNARY_EXPR:
		return p.formatTernary(n)

	case rl.K_CALL:
		return p.formatCall(n)

	case rl.K_VAR_PATH, rl.K_INDEXED_EXPR:
		return p.formatPath(n)

	case rl.K_PARENTHESIZED_EXPR:
		return p.formatParen(n)

	case rl.K_LIST:
		return p.formatList(n)

	case rl.K_MAP:
		return p.formatMap(n)

	default:
		return p.verbatim(n)
	}
}

// formatBinary renders `left op right` with single spaces around the operator -
// covers and/or, comparisons, in/not in, and arithmetic.
//
// [F20] binary operators get single spaces around them
func (p *printer) formatBinary(n *ts.Node) Doc {
	left := childByField(n, rl.F_LEFT)
	op := childByField(n, rl.F_OP)
	right := childByField(n, rl.F_RIGHT)
	if left == nil || op == nil || right == nil {
		return p.verbatim(n)
	}
	return concat(
		p.formatExpr(left),
		text(" "), text(p.nodeText(op)), text(" "),
		p.formatExpr(right),
	)
}

// formatUnary renders a prefix operator. Word operators (not) take a trailing
// space; symbolic ones (-, !) bind tight to their operand.
//
// [F21] unary: word ops (`not`) spaced, symbolic ops (`-`, `!`) tight
func (p *printer) formatUnary(n *ts.Node) Doc {
	op := childByField(n, rl.F_OP)
	arg := childByField(n, rl.F_ARG)
	if op == nil || arg == nil {
		return p.verbatim(n)
	}
	opText := p.nodeText(op)
	sep := ""
	if isWordOp(opText) {
		sep = " "
	}
	return concat(text(opText), text(sep), p.formatExpr(arg))
}

// formatTernary renders `cond ? a : b`.
//
// [F22] ternary: spaces around `?` and `:`
func (p *printer) formatTernary(n *ts.Node) Doc {
	cond := childByField(n, rl.F_CONDITION)
	tb := childByField(n, rl.F_TRUE_BRANCH)
	fb := childByField(n, rl.F_FALSE_BRANCH)
	if cond == nil || tb == nil || fb == nil {
		return p.verbatim(n)
	}
	return concat(
		p.formatExpr(cond),
		text(" ? "), p.formatExpr(tb),
		text(" : "), p.formatExpr(fb),
	)
}

// formatCall renders a function call: `f(a, b)`, wrapping one-arg-per-line with
// a trailing comma when it exceeds the line width. Positional and named args are
// selected by field name (the func, parens, and commas have other/no field, so
// they're skipped without special-casing).
//
// [F24] call args: tight parens, ", " between args
func (p *printer) formatCall(n *ts.Node) Doc {
	fn := childByField(n, rl.F_FUNC)
	var fnDoc Doc = text("")
	if fn != nil {
		fnDoc = p.formatExpr(fn)
	}

	var args []Doc
	for i, c := range childPtrs(n) {
		switch n.FieldNameForChild(uint32(i)) {
		case rl.F_ARG:
			args = append(args, p.formatExpr(c))
		case rl.F_NAMED_ARG:
			args = append(args, p.formatNamedArg(c))
		}
	}

	return concat(fnDoc, p.delimited(tLParen, tRParen, args, false))
}

// formatNamedArg formats a named call argument `key=val` tightly - no spaces
// around `=`. This distinguishes a call-site binding from an assignment
// statement (which does space its `=`).
//
// [F32] named call args bind tight: `f(key=val)`
func (p *printer) formatNamedArg(n *ts.Node) Doc {
	name := childByField(n, rl.F_NAME)
	value := childByField(n, rl.F_VALUE)
	if name == nil || value == nil {
		return p.verbatim(n)
	}
	return concat(text(p.nodeText(name)), text("="), p.formatExpr(value))
}

// formatParen renders `(expr)` tightly around its single inner expression.
// Parens are preserved exactly as written - never added or removed - so the
// author's grouping is respected.
//
// [F23] parentheses tight inside; never added or removed
func (p *printer) formatParen(n *ts.Node) Doc {
	for _, c := range namedChildrenOf(n) {
		return concat(text(tLParen), p.formatExpr(c), text(tRParen))
	}
	return p.verbatim(n)
}

// formatPath renders a variable path / postfix chain (`a.b.c`, `a[0]`,
// `obj.method(1)`) tightly, with no stray spaces, formatting interior index and
// call expressions.
//
// [F25] paths/postfix chains are tight (no spaces around `.` or `[]`)
func (p *printer) formatPath(n *ts.Node) Doc {
	var parts []Doc
	for i, c := range childPtrs(n) {
		field := n.FieldNameForChild(uint32(i))
		switch {
		case c.Kind() == tDot || c.Kind() == tLBracket || c.Kind() == tRBracket:
			parts = append(parts, text(c.Kind()))
		case c.Kind() == rl.K_SLICE:
			parts = append(parts, p.formatSlice(c))
		case field == rl.F_ROOT || field == rl.F_INDEXING || c.IsNamed():
			parts = append(parts, p.formatExpr(c))
		default:
			parts = append(parts, text(p.nodeText(c)))
		}
	}
	return concat(parts...)
}

// formatSlice renders `start:end` (and `::step` forms) tightly.
//
// [F26] slices are tight (no spaces around the slice colons)
func (p *printer) formatSlice(n *ts.Node) Doc {
	var parts []Doc
	for _, c := range childPtrs(n) {
		if c.Kind() == tColon {
			parts = append(parts, text(tColon))
		} else if c.IsNamed() {
			parts = append(parts, p.formatExpr(c))
		}
	}
	return concat(parts...)
}

// delimited renders a comma-separated list inside open/close delimiters, flat
// when it fits and one-item-per-line (with a trailing comma) when it doesn't.
// When pad is set, the flat form keeps a space just inside each delimiter
// (`{ a, b }`); otherwise it's tight (`[a, b]`, `f(a, b)`). The broken form is
// identical either way. An empty collection is always tight (`[]`, `{}`).
//
// [F29] over-width calls/collections wrap one item per line with a trailing comma
func (p *printer) delimited(open, close string, items []Doc, pad bool) Doc {
	if len(items) == 0 {
		return concat(text(open), text(close))
	}
	gap := softLine()
	if pad {
		gap = lineOrSpace()
	}
	return group(concat(
		text(open),
		indent(concat(
			gap,
			join(concat(text(tComma), lineOrSpace()), items),
			ifBreak(text(tComma), text("")),
		)),
		gap,
		text(close),
	))
}

// isWordOp reports whether an operator is alphabetic (and so needs spacing
// around it as a word) rather than symbolic.
func isWordOp(op string) bool {
	for _, r := range op {
		if !(r >= 'a' && r <= 'z') {
			return false
		}
	}
	return op != ""
}
