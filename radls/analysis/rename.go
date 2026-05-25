package analysis

import (
	"errors"
	"fmt"
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// radKeywords lists reserved words the Rad grammar accepts only in
// keyword positions. Renaming a symbol to any of these would produce
// an unparseable script. Kept in sync with tree-sitter-rad's grammar;
// the contextual keywords (`with`, `rad`, `request`, `display`,
// `confirm`, `unsafe`, `quiet`) are intentionally NOT here - the
// grammar aliases them as identifiers in non-keyword positions, so
// renaming to them is allowed (even if mildly unfortunate).
var radKeywords = map[string]bool{
	"if": true, "else": true, "while": true, "for": true, "in": true,
	"switch": true, "case": true, "yield": true,
	"break": true, "continue": true, "pass": true,
	"fn": true, "return": true, "defer": true, "errdefer": true,
	"and": true, "or": true, "not": true,
	"true": true, "false": true, "null": true,
	"args": true,
}

// ErrInvalidRenameTarget signals the cursor isn't on a renameable
// symbol - e.g. on whitespace, on a builtin (no source decl), or
// on a literal. The LSP client renders this as "Cannot rename" in
// the rename popup. The message is generic because the cursor's
// "what's wrong" depends on context (whitespace vs builtin vs
// unresolved); callers that need precision should wrap.
var ErrInvalidRenameTarget = errors.New("the symbol under the cursor cannot be renamed")

// ErrInvalidRenameName signals the requested new name isn't a
// legal Rad identifier. Wrapped by errInvalidRenameName(reason,
// name) at the call site so the LSP client surfaces *which* rule
// the input violated (digit start, bad character, reserved word).
var ErrInvalidRenameName = errors.New("not a valid Rad identifier")

// ErrRenameWouldCollide signals the new name is already in scope.
// Wrapped by errRenameCollision(name, ownerHint) so the message
// tells the user what they'd be colliding with - a same-scope
// local, a parent-scope binding, or an unloaded builtin.
var ErrRenameWouldCollide = errors.New("name is already in scope")

// errInvalidRenameName builds an LSP-visible message that explains
// *which* identifier rule the user broke. Wraps ErrInvalidRenameName
// so callers can still match the sentinel via errors.Is.
func errInvalidRenameName(name, reason string) error {
	return fmt.Errorf("%w: %s (%s)", ErrInvalidRenameName, name, reason)
}

// errRenameCollision builds an LSP-visible message that names the
// symbol the user would collide with. Wraps ErrRenameWouldCollide
// so callers can still match the sentinel via errors.Is.
func errRenameCollision(name, where string) error {
	return fmt.Errorf("%w: '%s' is already %s", ErrRenameWouldCollide, name, where)
}

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
	if reason, ok := validateRadIdentifier(newName); !ok {
		return nil, errInvalidRenameName(newName, reason)
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
	// Collision check has two arms. scopeCollision covers the loaded
	// scope chain (locals, params, fns the script already binds).
	// The builtin-set check covers builtins the script may not yet
	// reference - resolved.Builtin.Symbols is populated lazily as
	// names get used, so a script that never calls `print` would
	// otherwise pass a `rename x -> print` straight through and
	// silently shadow the builtin forever after.
	if where := scopeCollision(target.Scope, newName); where != "" {
		return nil, errRenameCollision(newName, where)
	}
	if _, isBuiltin := rts.FnSignaturesByName[newName]; isBuiltin {
		return nil, errRenameCollision(newName, "a built-in function")
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

// scopeCollision walks the symbol's scope (and its parents) for an
// existing binding under `name` and, on a hit, returns a short
// English phrase describing where it lives ("declared in this
// scope", "declared in an enclosing scope"). Empty string = no
// collision. Walking the parent chain catches the "you'd shadow
// an enclosing local" case in addition to "you'd collide with a
// same-scope local."
func scopeCollision(scope *check.Scope, name string) string {
	for cur := scope; cur != nil; cur = cur.Parent {
		if _, ok := cur.Symbols[name]; ok {
			if cur == scope {
				return "declared in this scope"
			}
			return "declared in an enclosing scope"
		}
	}
	return ""
}

// validateRadIdentifier reports whether s parses as a Rad
// identifier per the grammar and, on rejection, returns a short
// reason string so the caller can surface a specific message.
// Matches the shape the tree-sitter grammar accepts; we don't need
// to invoke the parser for the simple identifier rule, but we do
// need the keyword filter - without it, `rename x -> if` would
// emit textually-valid edits that produce an unparseable script.
//
// Empty string ok=true means "valid"; ok=false carries the reason.
func validateRadIdentifier(s string) (reason string, ok bool) {
	if s == "" {
		return "name is empty", false
	}
	first := s[0]
	if first >= '0' && first <= '9' {
		return "identifiers can't start with a digit", false
	}
	if !(first == '_' || (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return fmt.Sprintf("invalid first character '%c'", first), false
	}
	for i := 1; i < len(s); i++ {
		c := s[i]
		if !(c == '_' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return fmt.Sprintf("invalid character '%c'", c), false
		}
	}
	if radKeywords[s] {
		return "this is a reserved Rad keyword", false
	}
	return "", true
}
