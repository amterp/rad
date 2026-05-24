package analysis

import (
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// SemanticTokenType indexes into the legend below; the LSP wire
// format encodes token-type as the index, so the legend must stay
// in sync with these constants. Adding new types only appends to
// the end of the legend - never insert in the middle.
type SemanticTokenType int

const (
	TokenTypeFunction SemanticTokenType = iota
	TokenTypeParameter
	TokenTypeVariable
)

// SemanticTokensLegend is the shared client/server vocabulary.
// Exposed via the initialize response so the editor knows how to
// decode the indices in our emitted token data.
//
// We deliberately keep this list small. The LSP standard names
// run to 22 token types and 10 modifiers; emitting only the
// distinctions Rad's analyzer actually understands keeps the
// theming honest. We can grow this list as the analyzer's view
// gets richer (e.g. constants, enum members, type names).
func SemanticTokensLegend() lsp.SemanticTokensLegend {
	return lsp.SemanticTokensLegend{
		TokenTypes: []string{
			"function",
			"parameter",
			"variable",
		},
		TokenModifiers: []string{},
	}
}

// rawToken is the pre-encoding shape we collect during the AST
// walk. We emit (line, col, length, type) and let
// encodeSemanticTokens turn it into the spec's delta format.
type rawToken struct {
	line   int
	col    int
	length int
	ttype  SemanticTokenType
}

// SemanticTokens answers textDocument/semanticTokens/full: walk
// the resolved view and tag each identifier with its kind
// (function vs parameter vs variable). The editor uses this for
// finer-grained syntax highlighting than tree-sitter's
// CST-level coloring alone provides - e.g. distinguishing a
// builtin call from a local-variable read.
//
// Returns an empty data slice (not nil) when there's nothing to
// emit. The LSP spec lets us return null, but always-array makes
// the client's decode path simpler.
//
// Encoding: per LSP 3.17, tokens are sorted by position then
// delta-encoded as quintuples of uint. We do the sort and encode
// in one pass at the end so the AST walk can stay shape-blind.
func (s *State) SemanticTokens(snap *DocumentVersion) (*lsp.SemanticTokens, error) {
	out := &lsp.SemanticTokens{Data: []uint{}}
	if snap == nil || snap.ast == nil || snap.resolved == nil {
		return out, nil
	}

	tokens := collectSemanticTokens(snap)
	if len(tokens) == 0 {
		return out, nil
	}

	out.Data = encodeSemanticTokens(tokens, snap)
	return out, nil
}

// collectSemanticTokens walks every identifier and looks it up
// in resolved.Uses. Unresolved identifiers are skipped - the
// editor's tree-sitter highlighting handles them well enough
// as plain text, and emitting "variable" here would mis-color
// typos as if they were real bindings.
func collectSemanticTokens(snap *DocumentVersion) []rawToken {
	tokens := make([]rawToken, 0)
	walkAST(snap.ast, func(n rl.Node) {
		ident, ok := n.(*rl.Identifier)
		if !ok {
			return
		}
		sym, ok := snap.resolved.Uses[ident]
		if !ok || sym == nil {
			return
		}
		ttype, has := tokenTypeForSymbol(sym)
		if !has {
			return
		}
		s := ident.Span()
		tokens = append(tokens, rawToken{
			line:   s.StartRow,
			col:    s.StartCol,
			length: s.EndByte - s.StartByte,
			ttype:  ttype,
		})
	})
	return tokens
}

// tokenTypeForSymbol maps a SymbolKind to its semantic-token
// type, or false when the symbol shouldn't get a token of its
// own. We collapse builtins and hoisted user functions into the
// same "function" bucket - the editor doesn't render them
// distinctly today, and adding "builtin" as a separate type
// would lock us in before we have UX feedback.
func tokenTypeForSymbol(sym *check.Symbol) (SemanticTokenType, bool) {
	switch sym.Kind {
	case check.SymBuiltin, check.SymHoistedFn:
		return TokenTypeFunction, true
	case check.SymParam:
		return TokenTypeParameter, true
	case check.SymLocal, check.SymArg, check.SymCmdArg,
		check.SymLoopVar, check.SymWith:
		return TokenTypeVariable, true
	}
	return 0, false
}

// encodeSemanticTokens turns a flat token list into the LSP
// delta-encoded uint stream:
//
//	deltaLine, deltaStartChar, length, tokenType, tokenModifiers
//
// deltaLine is from the previous token's line; deltaStartChar
// resets to absolute when the line changes, else relative to
// the previous token on the same line. Length and column are
// translated through fromByteRange so multi-byte chars in
// UTF-16 clients don't render at the wrong width.
func encodeSemanticTokens(tokens []rawToken, snap *DocumentVersion) []uint {
	sort.Slice(tokens, func(i, j int) bool {
		if tokens[i].line != tokens[j].line {
			return tokens[i].line < tokens[j].line
		}
		return tokens[i].col < tokens[j].col
	})

	data := make([]uint, 0, len(tokens)*5)
	prevLine, prevCol := 0, 0
	for _, t := range tokens {
		// Translate the byte-column start (and the column of
		// `start + length`) into the negotiated encoding so the
		// editor highlights the correct visual extent. We compute
		// length as endCol-startCol in the negotiated encoding,
		// which is what the LSP wants: it measures length in
		// "character units" of the encoding the server advertised.
		idx := snap.lineIndex
		enc := snap.encoding
		startChar := idx.ByteColumnTo(t.line, t.col, enc)
		endChar := idx.ByteColumnTo(t.line, t.col+t.length, enc)
		length := endChar - startChar

		deltaLine := t.line - prevLine
		deltaCol := startChar
		if deltaLine == 0 {
			deltaCol = startChar - prevCol
		}
		data = append(data,
			uint(deltaLine),
			uint(deltaCol),
			uint(length),
			uint(t.ttype),
			0, // no modifiers yet
		)
		prevLine = t.line
		prevCol = startChar
	}
	return data
}
