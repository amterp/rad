package analysis

import (
	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// Package-scope AST utilities shared by every LSP feature handler.
// These used to live in hover.go because hover was the first
// consumer, but five other files now depend on them - keeping them
// here makes the package layout match what's actually shared.
//
// All helpers are pure: they don't depend on snapshot state and
// they don't mutate their inputs. Callers convert to/from the
// negotiated position encoding at the boundary (fromByteRange,
// toBytePos in state_actions.go).

// identifierAt walks the AST for the smallest Identifier whose
// span contains the given (line, byte-column) position. Cursor at
// the exact end column is treated as "on the identifier" - the
// cursor in `print|()` sits after the last char of `print` and
// users expect hover/goto-def to fire on `print` there.
//
// We do a full traversal rather than a span-pruning descent
// because the per-node Span isn't guaranteed to enclose every
// child's span (the converter sets each node's span independently
// and there are edge cases). It's O(n) over the AST; at our file
// sizes (sub-2k LOC) that's well below interactive-latency cost.
func identifierAt(root rl.Node, pos lsp.Pos) *rl.Identifier {
	var best *rl.Identifier
	rl.Walk(root, func(n rl.Node) {
		ident, ok := n.(*rl.Identifier)
		if !ok {
			return
		}
		if !spanContains(ident.Span(), pos) {
			return
		}
		if best == nil || spanSize(ident.Span()) < spanSize(best.Span()) {
			best = ident
		}
	})
	return best
}

// lookupSymbolForIdent finds the Symbol bound to an identifier,
// preferring the use-site index. Declarations (e.g. an `x` on the
// LHS of `x = 1`) are recorded in BOTH Decls and Uses by the
// binder so hover at the decl site finds the symbol; we still
// fall back to Decls just in case a future binder change drops
// the dual-registration for some kind.
func lookupSymbolForIdent(ident *rl.Identifier, resolved *check.Resolved) *check.Symbol {
	if sym, ok := resolved.Uses[ident]; ok && sym != nil {
		return sym
	}
	if sym, ok := resolved.Decls[ident]; ok && sym != nil {
		return sym
	}
	return nil
}

// spanContains reports whether (line, byteCol) sits inside span.
// The span is half-open at the start row but treats the end
// column as inclusive: a cursor sitting just after the last char
// of an identifier should still hover it (matches LSP-client
// expectation).
func spanContains(s rl.Span, pos lsp.Pos) bool {
	if pos.Line < s.StartRow || pos.Line > s.EndRow {
		return false
	}
	if pos.Line == s.StartRow && pos.Character < s.StartCol {
		return false
	}
	if pos.Line == s.EndRow && pos.Character > s.EndCol {
		return false
	}
	return true
}

// spanSize measures a span for "smallest containing" comparisons.
// We use byte length (EndByte - StartByte), which is unambiguous;
// row/column arithmetic would mis-rank multi-line spans.
func spanSize(s rl.Span) int {
	return s.EndByte - s.StartByte
}

// spanToRange converts an AST span (utf-8 byte columns) into an
// LSP Range still in byte coordinates. The caller is expected to
// push it through fromByteRange to land in the negotiated encoding.
func spanToRange(s rl.Span) lsp.Range {
	return lsp.Range{
		Start: lsp.Pos{Line: s.StartRow, Character: s.StartCol},
		End:   lsp.Pos{Line: s.EndRow, Character: s.EndCol},
	}
}
