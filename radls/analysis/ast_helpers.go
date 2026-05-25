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

// symbolAtPos generalises lookupSymbolForIdent to cover decl-site
// positions that DON'T sit on an *rl.Identifier node. Several Rad
// shapes store decl-site names as plain spans rather than identifier
// sub-nodes, and identifierAt can't reach them:
//
//   - fn name in `fn greet():` (FnDef.NameSpan)
//   - args-block name in `args: \n  name str` (ArgDecl.NameSpan)
//   - fn parameter name in `fn greet(name: str):`
//     (TypingFnParam.NameSpan, indexed via Resolved.ParamSymbols)
//   - for-loop variable in `for v in xs:`
//     (ForLoop.VarSpans, indexed via Resolved.ForLoopVars)
//
// The lookup prefers the identifier path - that's the common case
// and matches the existing behavior at use sites. Only when no
// identifier covers the cursor do we fall back to the decl-site
// walks; on a hit, the appropriate Resolved index yields the symbol
// the binder planted when the decl was bound.
func symbolAtPos(snap *DocumentVersion, pos lsp.Pos) *check.Symbol {
	if snap == nil || snap.ast == nil || snap.resolved == nil {
		return nil
	}
	if ident := identifierAt(snap.ast, pos); ident != nil {
		if sym := lookupSymbolForIdent(ident, snap.resolved); sym != nil {
			return sym
		}
	}
	var found *check.Symbol
	rl.Walk(snap.ast, func(n rl.Node) {
		if found != nil {
			return
		}
		switch nn := n.(type) {
		case *rl.FnDef:
			if nn.Name != "" && spanContains(nn.NameSpan, pos) {
				if sym, ok := snap.resolved.Decls[nn]; ok && sym != nil {
					found = sym
					return
				}
			}
			// Cursor on a parameter name: walk the typing.Params and
			// look up the symbol in ParamSymbols (keyed by owner).
			if syms := snap.resolved.ParamSymbols[nn]; syms != nil && nn.Typing != nil {
				found = paramSymbolAt(nn.Typing, syms, pos)
			}
		case *rl.Lambda:
			if syms := snap.resolved.ParamSymbols[nn]; syms != nil && nn.Typing != nil {
				found = paramSymbolAt(nn.Typing, syms, pos)
			}
		case *rl.ArgDecl:
			if nn.Name == "" || nn.NameSpan.EndByte == 0 || !spanContains(nn.NameSpan, pos) {
				return
			}
			if sym, ok := snap.resolved.Decls[nn]; ok && sym != nil {
				found = sym
			}
		case *rl.ForLoop:
			syms := snap.resolved.ForLoopVars[nn]
			if syms == nil {
				return
			}
			// VarSpans is parallel to Vars / syms; match by position.
			for i, span := range nn.VarSpans {
				if i >= len(syms) {
					break
				}
				if span.EndByte == 0 {
					continue
				}
				if spanContains(span, pos) {
					found = syms[i]
					return
				}
			}
			// Context name (`with ctx`): scope lookup by name since
			// the binder doesn't store the SymWith symbol in any
			// index. Acceptable because contexts are rare and the
			// fn body opens its own scope, so this read sees the
			// owning loop's scope first.
			if nn.Context != nil && nn.ContextSpan.EndByte != 0 && spanContains(nn.ContextSpan, pos) {
				// SymWith is declared in the loop's enclosing scope.
				// No direct AST-keyed map, but `Lookup` on the file
				// scope returns it; correctness rests on no other
				// symbol shadowing it between decl and the cursor.
				if sym := snap.resolved.File.Lookup(*nn.Context); sym != nil {
					found = sym
				}
			}
		}
	})
	return found
}

// paramSymbolAt returns the SymParam whose NameSpan covers pos, or
// nil. The Params and syms slices are in source order but may differ
// in length when some params have empty names (synthesised); we
// re-walk Params to keep the index alignment honest.
func paramSymbolAt(typing *rl.TypingFnT, syms []*check.Symbol, pos lsp.Pos) *check.Symbol {
	idx := 0
	for i := range typing.Params {
		p := &typing.Params[i]
		if p.Name == "" {
			continue
		}
		if idx >= len(syms) {
			break
		}
		if p.NameSpan.EndByte != 0 && spanContains(p.NameSpan, pos) {
			return syms[idx]
		}
		idx++
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
