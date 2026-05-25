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
// where `kind` tags the binding's origin (local, fn, arg, etc.)
// so users can tell at a glance whether a name is theirs or
// ambient. Builtins are an exception: their signature already
// makes "this is a function" obvious, so the kind tag would
// just add noise; we render builtins as `name: signature` with
// no prefix.
//
// Type comes from the strongest source available:
//  1. resolved.Decls / Uses -> Symbol
//  2. typeInfo.SymbolTypes[sym] if set (covers narrowed locals)
//  3. sym.Declared if pinned (typed-local, annotated param)
//  4. for SymBuiltin: FnSignaturesByName[name].Signature
//
// Returns "" when the identifier didn't resolve to a known
// symbol. The diagnostic squiggle already conveys that signal;
// adding an "(unresolved)" hover popup that echoes the typo
// back at the user is noise without information.
func formatIdentHover(ident *rl.Identifier, resolved *check.Resolved, info *check.TypeInfo) string {
	if resolved == nil {
		return ""
	}
	sym := lookupSymbolForIdent(ident, resolved)
	if sym == nil {
		return ""
	}

	typeStr := symbolTypeString(sym, info)
	if sym.Kind == check.SymBuiltin {
		// Signature is self-describing; the kind tag would
		// just repeat what `name(args) -> ret` already says.
		return fmt.Sprintf("```rad\n%s: %s\n```", sym.Name, typeStr)
	}
	kindLabel := symbolKindLabel(sym.Kind)
	return fmt.Sprintf("```rad\n(%s) %s: %s\n```", kindLabel, sym.Name, typeStr)
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
		// Internal builtins (`_rad_*`) are runtime plumbing - skip
		// the rich-signature path so they don't get an advertised
		// hover. Completion filters them too; the LSP surface should
		// stay consistent across both.
		if sig, ok := rts.FnSignaturesByName[sym.Name]; ok && !sig.IsInternal {
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
