package analysis

import (
	"github.com/amterp/rad/radls/lsp"
	"github.com/amterp/rad/rts/rl"
)

// zeroSpan marks "no decl site." Builtin Symbols leave DeclSpan
// at its zero value; we treat the full-struct zero (not just
// "byte range == 0..0") as the sentinel. A real top-of-file decl
// like `x = 1` still has EndByte > StartByte and a populated File,
// so it doesn't collide with the sentinel.
var zeroSpan = rl.Span{}

// Definition answers textDocument/definition: given a cursor on an
// identifier, return the Location of its declaration. Returns nil
// when there's nothing to jump to: the cursor isn't on an
// identifier, the name didn't resolve, or the symbol is a builtin
// (no source decl span to point at).
//
// Single-file today, single Location: Rad has no imports, so we
// never need to return multiple Locations. The LSP spec allows
// Location | Location[] | LocationLink[] | null; staying on Location
// keeps the wire shape simple and trivially forward-compatible.
//
// Builtins return nil rather than e.g. a synthetic stdlib URL.
// "Go to definition" on `print` doing nothing is much better than
// dumping the user into a generated file they can't edit; hover
// already shows them the signature, which is the actually-useful
// information.
func (s *State) Definition(snap *DocumentVersion, pos lsp.Pos) (*lsp.Location, error) {
	if snap == nil || snap.ast == nil || snap.resolved == nil {
		return nil, nil
	}

	bytePos := toBytePos(pos, snap)
	sym := symbolAtPos(snap, bytePos)
	if sym == nil {
		return nil, nil
	}

	// Builtins and any symbol whose DeclSpan is the zero value are
	// not navigable. The zero check guards future SymbolKinds where
	// we forgot to capture a span; better to return null than to
	// jump to (0,0) of the file.
	if sym.DeclSpan == (zeroSpan) {
		return nil, nil
	}

	r := fromByteRange(spanToRange(sym.DeclSpan), snap)
	return &lsp.Location{
		Uri:   snap.uri,
		Range: r,
	}, nil
}
