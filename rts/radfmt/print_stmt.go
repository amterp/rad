package radfmt

import (
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

const (
	kIfAlt   = "if_alt"
	kElseAlt = "else_alt"
	tElse    = "else"
)

// formatExprStmt formats a bare expression statement (e.g. a call).
func (p *printer) formatExprStmt(n *ts.Node) Doc {
	if e := childByField(n, rl.F_EXPR); e != nil {
		return p.formatExpr(e)
	}
	return p.verbatim(n)
}

// formatAssign formats `a = expr` and multi-assign `a, b = expr`.
//
// [F12] one space around `=`    [F13] multi-assign: ", " between targets
func (p *printer) formatAssign(n *ts.Node) Doc {
	var lefts []Doc
	var right *ts.Node
	for i, c := range childPtrs(n) {
		switch n.FieldNameForChild(uint32(i)) {
		case rl.F_LEFT, rl.F_LEFTS:
			lefts = append(lefts, p.formatExpr(c))
		case rl.F_RIGHT:
			right = c
		}
	}
	if len(lefts) == 0 || right == nil {
		return p.verbatim(n)
	}
	return concat(join(text(", "), lefts), text(" = "), p.formatExpr(right))
}

// formatTypedAssign formats a typed assignment `x: int = 1` - a space after the
// type colon and around `=`. The declared type is emitted verbatim for now
// (canonical spacing of `|`-union types is a noted follow-up). A trailing catch
// block (`... catch:`) is not yet handled, so fall back to verbatim when present
// rather than silently dropping it.
//
// [F31] typed assignment `x: int = 1`
func (p *printer) formatTypedAssign(n *ts.Node) Doc {
	if childByField(n, rl.F_CATCH) != nil {
		return p.verbatim(n)
	}
	left := childByField(n, rl.F_LEFT)
	typ := childByField(n, rl.F_DECLARED_TYPE)
	right := childByField(n, rl.F_RIGHT)
	if left == nil || typ == nil || right == nil {
		return p.verbatim(n)
	}
	return concat(
		text(p.nodeText(left)), text(": "), rawText(p.nodeText(typ)),
		text(" = "), p.formatExpr(right),
	)
}

// formatCompoundAssign formats `x += expr` (and -=, *=, /=, %=).
//
// [F14] one space around compound-assignment operators
func (p *printer) formatCompoundAssign(n *ts.Node) Doc {
	left := childByField(n, rl.F_LEFT)
	op := childByField(n, rl.F_OP)
	right := childByField(n, rl.F_RIGHT)
	if left == nil || op == nil || right == nil {
		return p.verbatim(n)
	}
	return concat(p.formatExpr(left), text(" "), text(p.nodeText(op)), text(" "), p.formatExpr(right))
}

// formatIncrDecr formats `i++` / `i--` tightly.
//
// [F15] increment/decrement bind tight (no inner space)
func (p *printer) formatIncrDecr(n *ts.Node) Doc {
	left := childByField(n, rl.F_LEFT)
	op := childByField(n, rl.F_OP)
	if left == nil || op == nil {
		return p.verbatim(n)
	}
	return concat(p.formatExpr(left), text(p.nodeText(op)))
}

// formatKeywordExpr formats statements that are a keyword optionally followed by
// an expression: `return`, `return expr`, `yield expr`.
//
// [F16] single space after return/yield keyword
func (p *printer) formatKeywordExpr(keyword string, n *ts.Node) Doc {
	for _, c := range namedChildrenOf(n) {
		return concat(text(keyword+" "), p.formatExpr(c))
	}
	return text(keyword)
}

// formatIf formats an if / else-if / else chain, each clause's body indented.
//
// [F17] if/else-if/else: header ends in `:`, body indented; `else if` on one line
func (p *printer) formatIf(n *ts.Node) Doc {
	var parts []Doc
	sawElse := false
	first := true
	for _, c := range childPtrs(n) {
		switch c.Kind() {
		case tElse:
			sawElse = true
		case kIfAlt:
			prefix := "if "
			if sawElse {
				prefix = "else if "
			}
			parts = append(parts, p.formatClause(prefix, c, !first))
			sawElse = false
			first = false
		case kElseAlt:
			parts = append(parts, p.formatClause("else", c, !first))
			sawElse = false
			first = false
		}
	}
	if len(parts) == 0 {
		return p.verbatim(n)
	}
	return concat(parts...)
}

// formatClause renders one if/else-if/else clause: a header line ending in `:`
// and an indented body. leadingBreak puts the clause on its own line after a
// preceding clause's body.
func (p *printer) formatClause(prefix string, alt *ts.Node, leadingBreak bool) Doc {
	header := Doc(text(prefix))
	if cond := childByField(alt, rl.F_CONDITION); cond != nil {
		header = concat(text(prefix), p.formatExpr(cond))
	}
	clause := concat(header, p.blockTail(alt))
	if leadingBreak {
		return concat(hardLine(), clause)
	}
	return clause
}

// formatFor formats `for x in expr:` (and `for i, x in expr:`), preserving the
// optional `with <ctx>` loop-context clause.
//
// [F18] for loop: ", " between loop vars, " in ", header ends in `:`
// [F43] for loop: keep the optional `with <ctx>` clause
func (p *printer) formatFor(n *ts.Node) Doc {
	lefts := childByField(n, rl.F_LEFTS)
	right := childByField(n, rl.F_RIGHT)
	if lefts == nil || right == nil {
		return p.verbatim(n)
	}
	header := concat(text("for "), p.formatForLefts(lefts), text(" in "), p.formatExpr(right))
	// The context is a bare identifier token (not a named expr), so emit its
	// source text directly - mirroring how formatForLefts handles loop vars.
	if ctx := childByField(n, rl.F_CONTEXT); ctx != nil {
		header = concat(header, text(" with "), text(p.nodeText(ctx)))
	}
	return concat(header, p.blockTail(n))
}

func (p *printer) formatForLefts(n *ts.Node) Doc {
	// The loop-variable identifiers carry field name "left" but are anonymous
	// tokens in the grammar, so select them by field rather than by IsNamed.
	var ids []Doc
	for i, c := range childPtrs(n) {
		if n.FieldNameForChild(uint32(i)) == rl.F_LEFT {
			ids = append(ids, text(p.nodeText(c)))
		}
	}
	return join(text(", "), ids)
}

// formatWhile formats `while cond:`.
//
// [F19] while loop: `while <cond>:`, body indented
func (p *printer) formatWhile(n *ts.Node) Doc {
	cond := childByField(n, rl.F_CONDITION)
	if cond == nil {
		return p.verbatim(n)
	}
	header := concat(text("while "), p.formatExpr(cond))
	return concat(header, p.blockTail(n))
}

// indentedBody renders a block body (the statements after a `:`) indented one
// level, each on its own line.
func (p *printer) indentedBody(items []*ts.Node) Doc {
	body := p.formatSeq(items)
	if body == nil {
		return text("")
	}
	return indent(concat(hardLine(), body))
}

// blockTail renders the `:` that opens a block, an optional comment that
// trailed the header on the same line (kept on the header line as a line-suffix
// rather than pulled into the body), and the indented body.
//
// [F11] a comment trailing a block header stays on the header line
func (p *printer) blockTail(n *ts.Node) Doc {
	headerComment, body := blockBody(n)
	tail := Doc(text(tColon))
	if headerComment != nil {
		tail = concat(tail, lineSuffix(concat(text(" "), text(p.nodeText(headerComment)))))
	}
	return concat(tail, p.indentedBody(body))
}

// blockBody splits a block opened by `:` into an optional same-line header
// comment and the remaining body items (statements and interleaved comments).
// A comment on the same row as the `:` documents the header, so it must stay on
// the header line rather than becoming the first body statement.
func blockBody(n *ts.Node) (headerComment *ts.Node, items []*ts.Node) {
	ch := childPtrs(n)
	colonIdx := -1
	for i, c := range ch {
		if c.Kind() == tColon {
			colonIdx = i
			break
		}
	}
	if colonIdx < 0 {
		return nil, nil
	}
	colon := ch[colonIdx]
	rest := ch[colonIdx+1:]
	if len(rest) > 0 && isComment(rest[0]) && sameRow(colon, rest[0]) {
		return rest[0], rest[1:]
	}
	return nil, rest
}
