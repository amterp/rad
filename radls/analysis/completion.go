package analysis

import (
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// Completion scope tiers. SortText values are leading-digit
// strings so the client's lexicographic sort on SortText puts
// "0" before "1" before "2", giving closer-scope items the top
// of the popup. Within each tier the Label tiebreaker keeps the
// list alphabetical, so the visual experience stays stable.
const (
	sortTierLocal    = "0"
	sortTierFile     = "1"
	sortTierBuiltin  = "2"
)

// buildCompletions populates `items` with every name reachable at
// the cursor: builtins, top-level fns and vars, args/cmd-args, and
// (when the cursor is inside a function or lambda body) that
// function's params plus any locals declared earlier in the body.
//
// Order discipline (the LSP client honours SortText before Label):
//
//   1. enclosing-fn params + earlier locals (tier "0")
//   2. file-scope: args, cmd-args, top-level fns + vars (tier "1")
//   3. builtins (tier "2")
//
// The popup the user sees has their local names at the top,
// file-scope next, builtins last - which matches what people
// actually want to type. Without SortText, alphabetical sort
// alone would bury a local `x` under whatever builtin starts
// with the same prefix.
//
// Dedupe with last-add-wins. The add order goes builtins ->
// file-scope -> enclosing-fn, so a name that exists in multiple
// scopes keeps the closest binding's metadata (Detail, SortText,
// Kind). This is also why the final sort must happen AFTER
// dedupe: the dedupe collapses entries by label but doesn't
// preserve insertion order across the map, so a separate sort
// is what guarantees stable ordering on the wire.
func buildCompletions(items *[]lsp.CompletionItem, snap *DocumentVersion, bytePos lsp.Pos) {
	if snap.ast == nil {
		// Even without an AST we can still offer builtins - they're
		// useful when the user has started a fresh file with a typo
		// and the parse failed.
		addBuiltinCompletions(items)
	} else {
		addBuiltinCompletions(items)
		addFileScopeCompletions(items, snap, bytePos)
		addEnclosingFnCompletions(items, snap, bytePos)
	}
	dedupCompletionsInPlace(items)
	// Sort runs unconditionally - including on the nil-AST path,
	// where without it the popup order would be whatever
	// FnSignaturesByName's map iteration produced this request.
	sort.SliceStable(*items, func(i, j int) bool {
		a, b := (*items)[i], (*items)[j]
		if a.SortText != b.SortText {
			return a.SortText < b.SortText
		}
		return a.Label < b.Label
	})
}

// addBuiltinCompletions offers every parsed builtin. The Detail is
// the signature so the editor's preview pane shows what the
// builtin expects without the user having to hover.
//
// Internal signatures (`_rad_*`) are filtered out - they exist for
// the runtime's own use and shouldn't show up in user-facing
// completion.
func addBuiltinCompletions(items *[]lsp.CompletionItem) {
	for name, sig := range rts.FnSignaturesByName {
		if sig.IsInternal {
			continue
		}
		*items = append(*items, lsp.CompletionItem{
			Label:    name,
			Kind:     lsp.CompletionKindFunction,
			Detail:   sig.Signature,
			SortText: sortTierBuiltin,
		})
	}
}

// addFileScopeCompletions walks SourceFile for top-level
// declarations: hoisted fns, top-level vars, args, cmd-args.
// We work directly off the AST rather than walking the resolved
// scope tree because the scope tree doesn't expose its children -
// the AST shape is the source of truth for "what's at top level."
//
// Position discipline: top-level Assigns aren't hoisted (the
// binder declares them at point of visit), so suggesting a var
// declared after the cursor would lead users to write code that
// runtime-errors with undefined-variable. We filter by
// spanBefore, matching the discipline in
// addEnclosingFnCompletions. FnDefs DO get hoisted by the
// binder so they're always in scope and we offer them
// regardless of position.
func addFileScopeCompletions(items *[]lsp.CompletionItem, snap *DocumentVersion, pos lsp.Pos) {
	file := snap.ast
	if file.Args != nil {
		for i := range file.Args.Decls {
			decl := &file.Args.Decls[i]
			*items = append(*items, lsp.CompletionItem{
				Label:    decl.Name,
				Kind:     lsp.CompletionKindVariable,
				Detail:   decl.TypeName,
				SortText: sortTierFile,
			})
		}
	}
	for _, cmd := range file.Cmds {
		for i := range cmd.Decls {
			decl := &cmd.Decls[i]
			*items = append(*items, lsp.CompletionItem{
				Label:    decl.Name,
				Kind:     lsp.CompletionKindVariable,
				Detail:   decl.TypeName,
				SortText: sortTierFile,
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
				Label:    n.Name,
				Kind:     lsp.CompletionKindFunction,
				Detail:   detail,
				SortText: sortTierFile,
			})
		case *rl.Assign:
			// Top-level vars aren't hoisted; suggesting one
			// declared after the cursor would offer a name the
			// runtime hasn't bound yet.
			if !spanBefore(n.Span(), pos) {
				continue
			}
			for _, target := range n.Targets {
				if ident, ok := target.(*rl.Identifier); ok {
					*items = append(*items, lsp.CompletionItem{
						Label: ident.Name,
						Kind:  lsp.CompletionKindVariable,
						Detail: localTypeString(
							ident, snap.resolved, snap.types,
						),
						SortText: sortTierFile,
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
			Label:    p.Name,
			Kind:     lsp.CompletionKindVariable,
			Detail:   typeStr,
			SortText: sortTierLocal,
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
					SortText: sortTierLocal,
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
