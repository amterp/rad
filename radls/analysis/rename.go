package analysis

import (
	"errors"
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// ErrInvalidRenameTarget signals the cursor isn't on a renameable
// symbol - e.g. on whitespace, on a builtin (no source decl), or
// on a literal. The LSP client renders this as "Cannot rename" in
// the rename popup.
var ErrInvalidRenameTarget = errors.New("the symbol under the cursor cannot be renamed")

// ErrInvalidRenameName signals the requested new name isn't a
// legal Rad identifier (empty, starts with a digit, contains
// non-identifier characters). The client surfaces the message in
// the rename dialog.
var ErrInvalidRenameName = errors.New("not a valid Rad identifier")

// ErrRenameWouldCollide signals the new name is already in scope -
// the rename would silently shadow or be shadowed by an existing
// symbol. Flagged here before the edit because letting it through
// would produce subtle behavioral changes in the script.
var ErrRenameWouldCollide = errors.New("name is already in scope")

// Rename answers textDocument/rename: produce a WorkspaceEdit
// covering every site that needs to change so the symbol under
// the cursor becomes `newName`. Single-file scope - Rad has no
// imports today, so cross-file renames don't apply.
//
// Returns:
//   - ErrInvalidRenameTarget when the cursor isn't on a
//     renameable symbol (whitespace, builtin, unresolved name).
//   - ErrInvalidRenameName when newName isn't a legal identifier.
//   - ErrRenameWouldCollide when the new name already binds
//     something in the target symbol's scope.
//   - A WorkspaceEdit with one TextEdit per site otherwise.
func (s *State) Rename(snap *DocumentVersion, pos lsp.Pos, newName string) (*lsp.WorkspaceEdit, error) {
	if snap == nil || snap.ast == nil || snap.resolved == nil {
		return nil, ErrInvalidRenameTarget
	}
	if !isValidRadIdentifier(newName) {
		return nil, ErrInvalidRenameName
	}

	bytePos := toBytePos(pos, snap)
	target := symbolAtPos(snap, bytePos)
	if target == nil {
		return nil, ErrInvalidRenameTarget
	}
	// Builtins have no source decl span; renaming `print` to
	// `say` would just silently produce an undefined-identifier
	// error at runtime. Reject before the edit.
	if target.Kind == check.SymBuiltin {
		return nil, ErrInvalidRenameTarget
	}
	// Same-name rename: no-op, treat as success with no edits.
	if target.Name == newName {
		empty := lsp.NewWorkspaceEdit()
		return &empty, nil
	}
	if scopeHasName(target.Scope, newName) {
		return nil, ErrRenameWouldCollide
	}

	// Collect every span the rename has to touch, then emit in
	// source order so the response is deterministic across
	// requests (Uses map iteration is non-deterministic).
	spans := []rl.Span{target.DeclSpan}
	seen := map[rl.Span]bool{target.DeclSpan: true}
	for node, sym := range snap.resolved.Uses {
		if sym != target {
			continue
		}
		if target.DefNode != nil && node == target.DefNode {
			continue
		}
		span := node.Span()
		if seen[span] {
			continue
		}
		seen[span] = true
		spans = append(spans, span)
	}
	sort.Slice(spans, func(i, j int) bool {
		a, b := spans[i], spans[j]
		if a.StartRow != b.StartRow {
			return a.StartRow < b.StartRow
		}
		return a.StartCol < b.StartCol
	})

	edit := lsp.NewWorkspaceEdit()
	for _, span := range spans {
		edit.AddEdit(snap.uri, fromByteRange(spanToRange(span), snap), newName)
	}
	return &edit, nil
}

// scopeHasName walks the symbol's scope (and its parents) for an
// existing binding under `name`. Walking the parent chain catches
// the "you'd shadow an enclosing local" case in addition to
// "you'd collide with a same-scope local."
func scopeHasName(scope *check.Scope, name string) bool {
	for cur := scope; cur != nil; cur = cur.Parent {
		if _, ok := cur.Symbols[name]; ok {
			return true
		}
	}
	return false
}

// isValidRadIdentifier reports whether s parses as a Rad
// identifier per the grammar: first byte is letter or underscore,
// remaining bytes are letter / digit / underscore. Matches the
// shape the tree-sitter grammar accepts; we don't need to invoke
// the parser for the simple identifier rule.
func isValidRadIdentifier(s string) bool {
	if s == "" {
		return false
	}
	first := s[0]
	if !(first == '_' || (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
