package analysis

import (
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/rl"
)

// DocumentSymbols answers textDocument/documentSymbol with the
// outline of a file: top-level functions, top-level variable
// declarations, the args block (with its declared args nested), and
// each cmd block (with its declared args nested). Returns an empty
// slice (NOT nil) when the file has no symbols - the LSP wire
// expects a JSON array.
//
// What counts as "top-level":
//
//   - FnDef in SourceFile.Stmts -> Function.
//   - First-assignment of an identifier in SourceFile.Stmts ->
//     Variable. Only the first assignment per name is emitted; later
//     re-assigns of the same name are not new outline entries.
//   - SourceFile.Args - the script's args: block - rendered as a
//     Namespace "args" with each ArgDecl as a Variable child.
//   - Each SourceFile.Cmds entry - rendered as a Module child group
//     named after the command, with its ArgDecls nested.
//
// We deliberately do NOT recurse into function bodies. Editors that
// want a deeper outline (per-statement, per-loop) can render local
// vars on demand via Hover/GotoDef; the outline is meant to be
// glanceable, not exhaustive.
func (s *State) DocumentSymbols(snap *DocumentVersion) ([]lsp.DocumentSymbol, error) {
	if snap == nil || snap.ast == nil {
		return []lsp.DocumentSymbol{}, nil
	}

	syms := make([]lsp.DocumentSymbol, 0)

	if snap.ast.Args != nil {
		syms = append(syms, argBlockSymbol(snap, snap.ast.Args))
	}
	for _, cmd := range snap.ast.Cmds {
		syms = append(syms, cmdBlockSymbol(snap, cmd))
	}

	seen := make(map[string]struct{})
	for _, stmt := range snap.ast.Stmts {
		switch n := stmt.(type) {
		case *rl.FnDef:
			syms = append(syms, fnDefSymbol(snap, n))
		case *rl.Assign:
			for _, target := range n.Targets {
				ident, ok := target.(*rl.Identifier)
				if !ok {
					continue
				}
				// First-decl wins: subsequent assignments of the same
				// name don't show up as separate outline entries.
				// Outline duplication makes the view harder to read,
				// not more informative.
				if _, dup := seen[ident.Name]; dup {
					continue
				}
				seen[ident.Name] = struct{}{}
				syms = append(syms, varSymbol(snap, ident, n.Span()))
			}
		}
	}

	// Render in source order, regardless of which top-level
	// construct appeared first in the AST traversal above.
	// Without this, args:/cmd: blocks always appear before any
	// fn/var even when the user wrote the fns first - the outline
	// would not match the file. Editors render symbols in the
	// order we return them.
	sort.SliceStable(syms, func(i, j int) bool {
		a, b := syms[i].Range.Start, syms[j].Range.Start
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Character < b.Character
	})

	return syms, nil
}

// argBlockSymbol renders the `args:` block as a namespace called
// "args" with each declared argument as a Variable child. The
// namespace kind is the closest LSP analogue to "group of related
// things" - VSCode renders it as a flat icon, which is fine since
// users recognize the args name.
func argBlockSymbol(snap *DocumentVersion, ab *rl.ArgBlock) lsp.DocumentSymbol {
	children := make([]lsp.DocumentSymbol, 0, len(ab.Decls))
	for i := range ab.Decls {
		decl := &ab.Decls[i]
		children = append(children, argDeclSymbol(snap, decl))
	}
	r := fromByteRange(spanToRange(ab.Span()), snap)
	return lsp.DocumentSymbol{
		Name:           "args",
		Kind:           lsp.SymbolKindNamespace,
		Range:          r,
		SelectionRange: r,
		Children:       children,
	}
}

// cmdBlockSymbol renders a command block (e.g. `build cmd:`) as a
// Module named after the command, with each of its arg decls
// nested as Variables. Two reasons for Module over Namespace: it
// renders with a distinct icon in most editors, and it more
// closely tracks the mental model ("a command is a subcommand of
// the script").
func cmdBlockSymbol(snap *DocumentVersion, cmd *rl.CmdBlock) lsp.DocumentSymbol {
	children := make([]lsp.DocumentSymbol, 0, len(cmd.Decls))
	for i := range cmd.Decls {
		decl := &cmd.Decls[i]
		children = append(children, argDeclSymbol(snap, decl))
	}
	r := fromByteRange(spanToRange(cmd.Span()), snap)
	return lsp.DocumentSymbol{
		Name:           cmd.Name,
		Kind:           lsp.SymbolKindModule,
		Range:          r,
		SelectionRange: r,
		Children:       children,
	}
}

// argDeclSymbol renders a single argument declaration as a
// Variable, with the type-string as Detail. SelectionRange covers
// just the name (which is what the editor highlights when the user
// clicks the outline entry), while Range covers the whole decl
// (name + type + default).
func argDeclSymbol(snap *DocumentVersion, decl *rl.ArgDecl) lsp.DocumentSymbol {
	whole := fromByteRange(spanToRange(decl.Span()), snap)
	return lsp.DocumentSymbol{
		Name:           decl.Name,
		Detail:         decl.TypeName,
		Kind:           lsp.SymbolKindVariable,
		Range:          whole,
		SelectionRange: whole,
	}
}

// fnDefSymbol renders a top-level function. SelectionRange covers
// just the function name (so the editor highlights the identifier
// when the user clicks the outline entry), Range covers the whole
// FnDef including the body. Detail carries a compact signature
// when annotations are present.
func fnDefSymbol(snap *DocumentVersion, fn *rl.FnDef) lsp.DocumentSymbol {
	whole := fromByteRange(spanToRange(fn.Span()), snap)
	sel := fromByteRange(spanToRange(fn.NameSpan), snap)
	detail := ""
	if fn.Typing != nil {
		detail = fn.Typing.Name()
	}
	return lsp.DocumentSymbol{
		Name:           fn.Name,
		Detail:         detail,
		Kind:           lsp.SymbolKindFunction,
		Range:          whole,
		SelectionRange: sel,
	}
}

// varSymbol renders a top-level variable binding. The enclosing
// Assign span makes the whole `x = expr` clickable in the outline;
// SelectionRange is just the name so the editor highlights the
// identifier, not the value.
func varSymbol(snap *DocumentVersion, ident *rl.Identifier, assignSpan rl.Span) lsp.DocumentSymbol {
	whole := fromByteRange(spanToRange(assignSpan), snap)
	sel := fromByteRange(spanToRange(ident.Span()), snap)
	return lsp.DocumentSymbol{
		Name:           ident.Name,
		Kind:           lsp.SymbolKindVariable,
		Range:          whole,
		SelectionRange: sel,
	}
}
