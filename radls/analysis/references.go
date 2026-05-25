package analysis

import (
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// References answers textDocument/references: given a cursor on an
// identifier, return every Location where that symbol is used in
// the document. The LSP context flag IncludeDeclaration toggles
// whether the declaration site itself is included.
//
// We compute on-demand from resolved.Uses rather than maintaining a
// pre-built reverse index. References is a click-driven feature
// (not background-streaming like diagnostics), so the cost of
// walking the Uses map once per request is fine - and it avoids
// growing every snapshot with a map most snapshots will never use.
// If the access pattern changes, the snapshot can grow a cache
// later without breaking callers.
//
// Returns an empty slice (not nil) when there are no references -
// the LSP wire expects an array.
func (s *State) References(snap *DocumentVersion, pos lsp.Pos, includeDecl bool) ([]lsp.Location, error) {
	if snap == nil || snap.ast == nil || snap.resolved == nil {
		return []lsp.Location{}, nil
	}

	bytePos := toBytePos(pos, snap)
	target := symbolAtPos(snap, bytePos)
	if target == nil {
		return []lsp.Location{}, nil
	}

	return collectReferences(snap, target, includeDecl), nil
}

// collectReferences walks the resolved indexes for every use of
// `target` and (optionally) the declaration site itself. The
// returned slice is sorted by start position so the editor's
// "references" panel renders in source order, which matches user
// expectation when scanning results.
//
// Why dedupe vs target.DefNode: the binder records the declaring
// identifier in BOTH Uses and Decls (so a hover at the decl site
// resolves through Uses like any other identifier). For
// references, that dual-registration would double-count the decl.
// We skip the use-entry whose node is the symbol's DefNode, then
// add the decl back in if the caller asked for it.
func collectReferences(snap *DocumentVersion, target *check.Symbol, includeDecl bool) []lsp.Location {
	uri := snap.uri
	out := make([]lsp.Location, 0)

	for node, sym := range snap.resolved.Uses {
		if sym != target {
			continue
		}
		if target.DefNode != nil && node == target.DefNode {
			continue // dual-registered decl identifier; handled below
		}
		out = append(out, locationFromSpan(uri, node.Span(), snap))
	}

	if includeDecl && target.DeclSpan != zeroSpan {
		out = append(out, locationFromSpan(uri, target.DeclSpan, snap))
	}

	// Sort by start position (line, then character). Map iteration
	// in Go is non-deterministic, so this is the difference between
	// "the editor shows results in source order" and "the editor
	// shows them in whatever order the map iterated this time."
	sort.Slice(out, func(i, j int) bool {
		a, b := out[i].Range.Start, out[j].Range.Start
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Character < b.Character
	})

	return out
}

// locationFromSpan builds an LSP Location from an AST span in the
// negotiated encoding. Kept here (not in hover/definition) so this
// file is the single place that knows about Span-to-Location for
// the references case, even though the conversion is trivial - the
// encoding-translation step is easy to forget.
func locationFromSpan(uri string, span rl.Span, snap *DocumentVersion) lsp.Location {
	return lsp.Location{
		Uri:   uri,
		Range: fromByteRange(spanToRange(span), snap),
	}
}
