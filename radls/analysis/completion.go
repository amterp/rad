package analysis

import (
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// buildCompletions populates `items` with every name reachable at
// the cursor: builtins, top-level fns and vars, args/cmd-args, and
// (when the cursor is inside a function or lambda body) that
// function's params plus any locals declared earlier in the body.
//
// Single dedupe pass at the end: a name might be in scope from
// several angles (e.g. an `args:` arg `name` and a builtin called
// `name`), and we want one entry per unique label so the editor
// list doesn't show duplicates. Deeper/closer bindings win the
// detail slot since that's the binding the cursor would actually
// resolve to.
//
// Sorting: alphabetical by label so the popup is stable across
// requests. Some clients re-sort by their own scoring (recency,
// fuzzy-rank) but starting alphabetical is friendlier than the
// random order that map iteration would produce.
func buildCompletions(items *[]lsp.CompletionItem, snap *DocumentVersion, bytePos lsp.Pos) {
	if snap.ast == nil {
		// Even without an AST we can still offer builtins - they're
		// useful when the user has started a fresh file with a typo
		// and the parse failed.
		addBuiltinCompletions(items)
		dedupCompletionsInPlace(items)
		return
	}

	addBuiltinCompletions(items)
	addFileScopeCompletions(items, snap)
	addEnclosingFnCompletions(items, snap, bytePos)
	dedupCompletionsInPlace(items)
	sort.SliceStable(*items, func(i, j int) bool {
		return (*items)[i].Label < (*items)[j].Label
	})
}

// addBuiltinCompletions offers every parsed builtin. The Detail is
// the signature so the editor's preview pane shows what the
// builtin expects without the user having to hover.
func addBuiltinCompletions(items *[]lsp.CompletionItem) {
	for name, sig := range rts.FnSignaturesByName {
		*items = append(*items, lsp.CompletionItem{
			Label:  name,
			Kind:   lsp.CompletionKindFunction,
			Detail: sig.Signature,
		})
	}
}

// addFileScopeCompletions walks SourceFile for top-level
// declarations: hoisted fns, top-level vars, args, cmd-args.
// We work directly off the AST rather than walking the resolved
// scope tree because the scope tree doesn't expose its children -
// the AST shape is the source of truth for "what's at top level."
func addFileScopeCompletions(items *[]lsp.CompletionItem, snap *DocumentVersion) {
	file := snap.ast
	if file.Args != nil {
		for i := range file.Args.Decls {
			decl := &file.Args.Decls[i]
			*items = append(*items, lsp.CompletionItem{
				Label:  decl.Name,
				Kind:   lsp.CompletionKindVariable,
				Detail: decl.TypeName,
			})
		}
	}
	for _, cmd := range file.Cmds {
		for i := range cmd.Decls {
			decl := &cmd.Decls[i]
			*items = append(*items, lsp.CompletionItem{
				Label:  decl.Name,
				Kind:   lsp.CompletionKindVariable,
				Detail: decl.TypeName,
			})
		}
	}
	for _, stmt := range file.Stmts {
		switch n := stmt.(type) {
		case *rl.FnDef:
			detail := ""
			if n.Typing != nil {
				detail = n.Typing.Name()
			}
			*items = append(*items, lsp.CompletionItem{
				Label:  n.Name,
				Kind:   lsp.CompletionKindFunction,
				Detail: detail,
			})
		case *rl.Assign:
			for _, target := range n.Targets {
				if ident, ok := target.(*rl.Identifier); ok {
					*items = append(*items, lsp.CompletionItem{
						Label: ident.Name,
						Kind:  lsp.CompletionKindVariable,
						Detail: localTypeString(
							ident, snap.resolved, snap.types,
						),
					})
				}
			}
		}
	}
}

// addEnclosingFnCompletions adds params and body-locals of the
// smallest fn/lambda whose body contains the cursor. The cursor
// at file scope hits the no-enclosing case and we skip cleanly.
//
// We walk the AST again rather than threading the binder's scope
// stack here because the binder doesn't expose a position-keyed
// scope index. The cost is one O(n) walk per request, which is
// noise at our file sizes; building and threading a scope index
// is real complexity to be paid only if benchmarks demand it.
func addEnclosingFnCompletions(items *[]lsp.CompletionItem, snap *DocumentVersion, pos lsp.Pos) {
	owner, body := enclosingCallable(snap.ast, pos)
	if owner == nil {
		return
	}
	for _, p := range paramsOf(owner) {
		typeStr := ""
		if p.Type != nil {
			typeStr = (*p.Type).Name()
		}
		*items = append(*items, lsp.CompletionItem{
			Label:  p.Name,
			Kind:   lsp.CompletionKindVariable,
			Detail: typeStr,
		})
	}
	for _, stmt := range body {
		assign, ok := stmt.(*rl.Assign)
		if !ok {
			continue
		}
		// Only suggest locals declared BEFORE the cursor. Suggesting
		// a name that isn't yet in scope would lead users to
		// reference variables the runtime hasn't seen.
		if !spanBefore(assign.Span(), pos) {
			continue
		}
		for _, target := range assign.Targets {
			if ident, ok := target.(*rl.Identifier); ok {
				*items = append(*items, lsp.CompletionItem{
					Label: ident.Name,
					Kind:  lsp.CompletionKindVariable,
					Detail: localTypeString(
						ident, snap.resolved, snap.types,
					),
				})
			}
		}
	}
}

// enclosingCallable returns the smallest FnDef/Lambda whose body
// span covers the cursor, plus the body slice for body-local
// extraction. Returns (nil, nil) when the cursor is at file scope.
func enclosingCallable(root rl.Node, pos lsp.Pos) (rl.Node, []rl.Node) {
	var owner rl.Node
	var body []rl.Node
	rl.Walk(root, func(n rl.Node) {
		switch nn := n.(type) {
		case *rl.FnDef:
			if spanContains(nn.Span(), pos) {
				if owner == nil || spanSize(nn.Span()) < ownerSize(owner) {
					owner, body = nn, nn.Body
				}
			}
		case *rl.Lambda:
			if spanContains(nn.Span(), pos) {
				if owner == nil || spanSize(nn.Span()) < ownerSize(owner) {
					owner, body = nn, nn.Body
				}
			}
		}
	})
	return owner, body
}

// ownerSize sidesteps Go's lack of variant dispatch on
// (*FnDef|*Lambda).Span(). They both have Span() but we can't
// call it through rl.Node here because Node is an interface and
// the values are concrete - a type switch is the idiomatic path.
func ownerSize(owner rl.Node) int {
	return spanSize(owner.Span())
}

// paramsOf returns the param list of an FnDef or Lambda; empty
// for anything else (defensive: caller already type-switched but
// guarding here keeps the helper safe to extend).
func paramsOf(owner rl.Node) []rl.TypingFnParam {
	switch n := owner.(type) {
	case *rl.FnDef:
		if n.Typing != nil {
			return n.Typing.Params
		}
	case *rl.Lambda:
		if n.Typing != nil {
			return n.Typing.Params
		}
	}
	return nil
}

// spanBefore reports whether `s` ends strictly before `pos`. Used
// to filter out locals declared after the cursor - those aren't
// in scope yet from the user's perspective.
func spanBefore(s rl.Span, pos lsp.Pos) bool {
	if s.EndRow < pos.Line {
		return true
	}
	if s.EndRow == pos.Line && s.EndCol <= pos.Character {
		return true
	}
	return false
}

// localTypeString picks the best type string for a local being
// completed. We route through the resolved/types indexes when
// available because the type checker's narrowing-aware view is
// strictly more informative than the raw assignment RHS would be.
func localTypeString(ident *rl.Identifier, resolved *check.Resolved, info *check.TypeInfo) string {
	if resolved == nil {
		return ""
	}
	sym := lookupSymbolForIdent(ident, resolved)
	if sym == nil {
		return ""
	}
	if info != nil {
		if t, ok := info.SymbolTypes[sym]; ok && t != nil {
			return t.Name()
		}
	}
	if sym.Declared != nil {
		return sym.Declared.Name()
	}
	return ""
}

// dedupCompletionsInPlace collapses duplicate labels to a single
// entry, keeping the LAST-added one. Order matters here: we add
// builtins first, then file scope, then enclosing-fn scope - so
// "last wins" means the closest binding's detail survives, which
// is the binding the cursor would actually resolve to.
func dedupCompletionsInPlace(items *[]lsp.CompletionItem) {
	by := make(map[string]lsp.CompletionItem, len(*items))
	for _, it := range *items {
		by[it.Label] = it
	}
	out := (*items)[:0]
	for _, it := range by {
		out = append(out, it)
	}
	*items = out
}
