package radfmt

import (
	"strings"

	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// printer turns a CST into a Doc. It holds the original source so it can emit
// verbatim spans for constructs that don't yet have a dedicated formatter.
type printer struct {
	src string
}

// format dispatches on node kind. As construct formatters are implemented they
// get their own case; everything else falls through to verbatim, which is always
// safe (it re-emits the node's exact original source).
//
// Deliberately not yet formatted (they fall through to verbatim and are emitted
// unchanged - structurally safe, just not canonicalized):
//   - strings: quote-style normalization and interpolation reflow
//   - arg_block, cmd_block, rad_block
//   - fn_named / fn_lambda
//   - switch_stmt, defer_block, shell_stmt
//   - list_comprehension
//
// To add one: handle its kind here (or in formatExpr for expressions), build a
// Doc from its fields, and add a snapshot case. The structural-equivalence guard
// will reject any formatter that changes what the code parses to.
func (p *printer) format(node *ts.Node) Doc {
	switch node.Kind() {
	case rl.K_SOURCE_FILE:
		return p.formatSourceFile(node)

	// Statements.
	case rl.K_EXPR_STMT:
		return p.formatExprStmt(node)
	case rl.K_ASSIGN:
		return p.formatAssign(node)
	case rl.K_TYPED_ASSIGN:
		return p.formatTypedAssign(node)
	case rl.K_COMPOUND_ASSIGN:
		return p.formatCompoundAssign(node)
	case rl.K_INCR_DECR:
		return p.formatIncrDecr(node)
	case rl.K_RETURN_STMT:
		return p.formatKeywordExpr("return", node)
	case rl.K_YIELD_STMT:
		return p.formatKeywordExpr("yield", node)
	case rl.K_IF_STMT:
		return p.formatIf(node)
	case rl.K_FOR_LOOP:
		return p.formatFor(node)
	case rl.K_WHILE_LOOP:
		return p.formatWhile(node)

	// Expression nodes, in case format() is called on one directly.
	case rl.K_EXPR, rl.K_TERNARY_EXPR, rl.K_OR_EXPR, rl.K_AND_EXPR,
		rl.K_COMPARE_EXPR, rl.K_ADD_EXPR, rl.K_MULT_EXPR, rl.K_UNARY_EXPR,
		rl.K_CALL, rl.K_VAR_PATH, rl.K_INDEXED_EXPR, rl.K_PARENTHESIZED_EXPR,
		rl.K_LIST, rl.K_MAP, rl.K_STRING, rl.K_IDENTIFIER, rl.K_INT,
		rl.K_FLOAT, rl.K_BOOL, rl.K_NULL:
		return p.formatExpr(node)

	default:
		return p.verbatim(node)
	}
}

// formatSourceFile lays out the top-level statement sequence: each statement on
// its own line(s), single blank lines preserved (multiples collapsed, none at
// file edges), comments placed as leading/standalone/trailing, and exactly one
// trailing newline. Individual statements are formatted by format(); any not yet
// handled fall through to verbatim.
//
// [F3] exactly one trailing newline at end of file
func (p *printer) formatSourceFile(node *ts.Node) Doc {
	body := p.formatSeq(childPtrs(node))
	if body == nil {
		return text("")
	}
	return concat(body, hardLine())
}

// formatSeq renders an ordered sequence of statements and comments (the body of
// a file or block) with canonical separators:
//   - one hardline between adjacent items,
//   - a second hardline where the source had at least one blank line,
//   - a same-line comment following a statement attaches as a trailing
//     line-suffix rather than starting a new line.
//
// It returns nil for an empty sequence so callers can special-case emptiness.
//
// [F6] collapse 2+ blank lines to one    [F7] strip blanks at block/file edges
// [F8] preserve a single blank line       [F9] standalone comment keeps its line
// [F10] trailing same-line comment stays on the statement's line
func (p *printer) formatSeq(items []*ts.Node) Doc {
	if len(items) == 0 {
		return nil
	}

	var parts []Doc
	emitted := false
	var prev *ts.Node

	for i := 0; i < len(items); i++ {
		item := items[i]

		if emitted {
			parts = append(parts, hardLine()) // [F8] one hardline between items
			if blankBetween(prev, item) {
				parts = append(parts, hardLine()) // [F6][F8] at most one blank line
			}
		}

		if isComment(item) {
			parts = append(parts, text(p.nodeText(item))) // [F9] standalone comment
			prev = item
		} else {
			doc := p.format(item)
			// Fold a trailing same-line comment into this statement's line. [F10]
			if i+1 < len(items) && isComment(items[i+1]) && sameRow(item, items[i+1]) {
				c := items[i+1]
				doc = concat(doc, lineSuffix(concat(text(" "), text(p.nodeText(c)))))
				prev = c
				i++
			} else {
				prev = item
			}
			parts = append(parts, doc)
		}
		emitted = true
	}

	return concat(parts...)
}

// childPtrs returns stable pointers to a node's children (named, anonymous, and
// comment extras) in source order, suitable for sequence walking.
func childPtrs(n *ts.Node) []*ts.Node {
	cs := children(n)
	out := make([]*ts.Node, len(cs))
	for i := range cs {
		out[i] = &cs[i]
	}
	return out
}

// verbatim re-emits a node's exact source span, minus any trailing newline -
// the statement sequencer owns all inter-statement spacing, so a node's own
// trailing newline must not be emitted or it doubles up into a blank line. Inner
// comments and original layout are preserved because they're part of the span;
// embedded newlines render as literal (no-indent) breaks so the column counter
// stays correct.
//
// This is also the path by which deliberately-untouched constructs are emitted:
// [F34] the shebang line and [F35] the `--- ... ---` file header are preserved
// exactly (along with any construct not yet given a dedicated formatter).
func (p *printer) verbatim(node *ts.Node) Doc {
	raw := strings.TrimRight(p.src[node.StartByte():node.EndByte()], "\n")
	return rawText(raw)
}

// rawText turns arbitrary source text (possibly multi-line) into a Doc that
// renders byte-for-byte identically: each line becomes Text, separated by
// literal (no-indent) line breaks.
func rawText(s string) Doc {
	if !strings.Contains(s, "\n") {
		return text(s)
	}
	lines := strings.Split(s, "\n")
	parts := make([]Doc, 0, len(lines)*2-1)
	for i, ln := range lines {
		if i > 0 {
			parts = append(parts, literalLine())
		}
		if ln != "" {
			parts = append(parts, text(ln))
		}
	}
	return concat(parts...)
}
