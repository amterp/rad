package analysis

import (
	"fmt"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// Hover answers textDocument/hover against a fixed document snapshot.
// We find the smallest Identifier whose span covers the cursor, look
// up its Symbol in the resolved view, and format the symbol's type
// (declared annotation if any, else what the type checker inferred).
// Returns nil when the cursor isn't on a hoverable thing - the LSP
// spec lets us return null and the client will simply show nothing.
//
// Why identifier-only for v1: the value gradient is steep. Hovering a
// name is what users want 90% of the time; expression-result hover
// (e.g. on a binop or call) is nice-to-have and lights up trivially
// once we route through TypeInfo.ExprTypes here. We'll grow it as
// users notice it's missing.
func (s *State) Hover(snap *DocumentVersion, pos lsp.Pos) (*lsp.Hover, error) {
	if snap == nil || snap.ast == nil {
		return nil, nil
	}

	bytePos := toBytePos(pos, snap)
	ident := identifierAt(snap.ast, bytePos)
	if ident == nil {
		return nil, nil
	}

	contents := formatIdentHover(ident, snap.resolved, snap.types)
	if contents == "" {
		return nil, nil
	}

	r := fromByteRange(spanToRange(ident.Span()), snap)
	return &lsp.Hover{
		Contents: lsp.MarkupContent{
			Kind:  lsp.MarkupMarkdown,
			Value: contents,
		},
		Range: &r,
	}, nil
}

// identifierAt walks the AST for the smallest Identifier whose span
// contains the given (line, byte-column) position. Cursor at the
// exact end column is treated as "on the identifier" - the cursor in
// `print|()` sits after the last char of `print` and users expect
// hover on `print` there.
//
// We do a full traversal rather than a span-pruning descent because
// the per-node Span isn't guaranteed to enclose every child's span
// (the converter sets each node's span independently and there are
// edge cases). It's O(n) over the AST; for sub-2k-LOC scripts that's
// well below interactive-latency cost.
func identifierAt(root rl.Node, pos lsp.Pos) *rl.Identifier {
	var best *rl.Identifier
	walkAST(root, func(n rl.Node) {
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

// walkAST is a local copy of the check-package helper - we don't want
// to pull check's internals here just for a tree walk. Hover lives in
// the analysis package, which already imports check for the indexes
// it consumes, but the visitor is plain AST traversal.
func walkAST(node rl.Node, visit func(rl.Node)) {
	if node == nil {
		return
	}
	visit(node)
	for _, child := range node.Children() {
		walkAST(child, visit)
	}
}

// spanContains reports whether (line, byteCol) sits inside span. The
// span is half-open at the start row but treats the end column as
// inclusive: a cursor sitting just after the last char of an
// identifier should still hover it (matches LSP-client expectation).
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

// spanSize measures a span for "smallest containing" comparisons. We
// use byte length (EndByte - StartByte), which is unambiguous; row/
// column arithmetic would mis-rank multi-line spans.
func spanSize(s rl.Span) int {
	return s.EndByte - s.StartByte
}

// spanToRange converts an AST span (utf-8 byte columns) into an LSP
// Range still in byte coordinates. The caller is expected to push it
// through fromByteRange to land in the negotiated encoding.
func spanToRange(s rl.Span) lsp.Range {
	return lsp.Range{
		Start: lsp.Pos{Line: s.StartRow, Character: s.StartCol},
		End:   lsp.Pos{Line: s.EndRow, Character: s.EndCol},
	}
}

// formatIdentHover renders the markdown body for a hover on an
// identifier. Returns "" when there's nothing useful to say (e.g.
// the identifier didn't resolve to a known symbol). The empty-
// return contract lets the caller short-circuit to a null hover.
//
// Format - Rust-rust-analyzer flavoured:
//
//	```rad
//	(kind) name: type
//	```
//
// where `kind` tags the binding's origin (local, fn, builtin, etc.)
// so users can tell at a glance whether a name is theirs or
// ambient. Type comes from the strongest source available:
//  1. resolved.Decls / Uses -> Symbol
//  2. typeInfo.SymbolTypes[sym] if set (covers narrowed locals)
//  3. sym.Declared if pinned (typed-local, annotated param)
//  4. for SymBuiltin: FnSignaturesByName[name].Signature
//
// Falls back to "?" when we have a symbol but no recoverable type;
// "(unresolved)" when we don't even have a symbol.
func formatIdentHover(ident *rl.Identifier, resolved *check.Resolved, info *check.TypeInfo) string {
	if resolved == nil {
		return ""
	}
	sym := lookupSymbolForIdent(ident, resolved)
	if sym == nil {
		return fmt.Sprintf("```rad\n(unresolved) %s\n```", ident.Name)
	}

	kindLabel := symbolKindLabel(sym.Kind)
	typeStr := symbolTypeString(sym, info)
	return fmt.Sprintf("```rad\n(%s) %s: %s\n```", kindLabel, sym.Name, typeStr)
}

// lookupSymbolForIdent finds the Symbol bound to an identifier,
// preferring the use-site index. Declarations (e.g. an `x` on the
// LHS of `x = 1`) are sometimes recorded in Decls but not Uses; we
// fall back to Decls so hover works on the declaration itself.
func lookupSymbolForIdent(ident *rl.Identifier, resolved *check.Resolved) *check.Symbol {
	if sym, ok := resolved.Uses[ident]; ok && sym != nil {
		return sym
	}
	if sym, ok := resolved.Decls[ident]; ok && sym != nil {
		return sym
	}
	return nil
}

// symbolKindLabel maps a SymbolKind to a short tag for the hover
// header. The labels are chosen to mirror what a user would call
// these in conversation, not the internal enum names.
func symbolKindLabel(k check.SymbolKind) string {
	switch k {
	case check.SymBuiltin:
		return "builtin"
	case check.SymHoistedFn:
		return "fn"
	case check.SymArg:
		return "arg"
	case check.SymCmdArg:
		return "cmd arg"
	case check.SymParam:
		return "param"
	case check.SymLocal:
		return "local"
	case check.SymLoopVar:
		return "loop var"
	case check.SymWith:
		return "with"
	}
	return "symbol"
}

// symbolTypeString picks the best available textual rendering for a
// symbol's type. The order prefers the most-specific information
// the analyzer has: flow-sensitive inferred type > declared
// annotation > raw builtin signature. Anything that falls through
// renders as "?" so the hover is still non-empty.
func symbolTypeString(sym *check.Symbol, info *check.TypeInfo) string {
	if sym.Kind == check.SymBuiltin {
		if sig, ok := rts.FnSignaturesByName[sym.Name]; ok {
			return sig.Signature
		}
	}
	if info != nil {
		if t, ok := info.SymbolTypes[sym]; ok && t != nil {
			return t.Name()
		}
	}
	if sym.Declared != nil {
		return sym.Declared.Name()
	}
	return "?"
}
