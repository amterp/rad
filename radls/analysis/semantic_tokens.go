package analysis

import (
	"sort"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
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
	TokenTypeType
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
			"type",
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

// collectSemanticTokens walks the AST for identifiers and emits
// a token for each. Three sources:
//
//  1. Identifier nodes that resolve through resolved.Uses. This
//     catches call sites, var reads, and the dual-registered
//     decl identifier on `x = 1`.
//  2. FnDef name positions. The binder declares hoisted fns at
//     the AST node (not at an Identifier), so the name token
//     at `fn greet():` has no Uses entry; without this branch
//     it would not be coloured at the decl site even though
//     every call to `greet()` is.
//  3. ArgDecl name positions inside the `args:` / cmd `args:`
//     blocks. Same shape as FnDef: the args declaration sits in
//     the AST as an ArgDecl node rather than an Identifier, so
//     the decl-site name needs its own emit to match how every
//     use of the arg is tagged via path 1.
//
// Fn param-name positions at the decl site are still uncoloured.
// The AST's TypingFnParam carries no per-name span; emitting
// would need a converter+AST extension. Param references inside
// the body do get tokens via path 1.
//
// Unresolved identifiers are skipped - the editor's tree-sitter
// highlighting handles them well enough as plain text, and
// emitting "variable" here would mis-color typos as if they
// were real bindings.
func collectSemanticTokens(snap *DocumentVersion) []rawToken {
	tokens := make([]rawToken, 0)
	rl.Walk(snap.ast, func(n rl.Node) {
		switch nn := n.(type) {
		case *rl.Identifier:
			sym, ok := snap.resolved.Uses[nn]
			if !ok || sym == nil {
				return
			}
			ttype, has := tokenTypeForSymbol(sym)
			if !has {
				return
			}
			tokens = append(tokens, tokenFromSpan(nn.Span(), ttype))
		case *rl.FnDef:
			// Skip anonymous or zero-name forms; NameSpan would be
			// the whole node and emitting that would over-paint.
			if nn.Name == "" {
				return
			}
			tokens = append(tokens, tokenFromSpan(nn.NameSpan, TokenTypeFunction))
		case *rl.ArgDecl:
			if nn.Name == "" || nn.NameSpan.EndByte == 0 {
				return
			}
			tokens = append(tokens, tokenFromSpan(nn.NameSpan, TokenTypeParameter))
		}
	})
	if snap.tree != nil {
		collectTypeTokens(snap.tree.Root(), &tokens)
	}
	return tokens
}

// typeKeywordLengths maps a CST node kind to the byte length of the
// leading keyword that should carry the `type` semantic token. For
// nodes whose entire span IS the keyword (string_type = "str") the
// length is the whole node. For composite forms (`int[]`) we only
// want the leading word, not the brackets.
var typeKeywordLengths = map[string]int{
	"string_type":      -1, // whole span ("str")
	"int_type":         -1, // ("int")
	"float_type":       -1, // ("float")
	"bool_type":        -1, // ("bool")
	"error_type":       -1, // ("error")
	"any_type":         -1, // ("any")
	"void_type":        -1, // ("void")
	"string_list_type": 3,  // "str" prefix of "str[]"
	"int_list_type":    3,  // "int" prefix of "int[]"
	"float_list_type":  5,  // "float" prefix
	"bool_list_type":   4,  // "bool" prefix
}

// collectTypeTokens walks the CST and emits a `type` semantic token
// for every type-annotation keyword. The CST is the right level for
// this: the AST's TypingT shapes carry no source positions, but the
// CST has byte-exact spans on each `string_type` / `int_type` /
// etc. node. We deliberately limit emission to the keyword bytes
// (not the whole composite) so the surrounding punctuation - `[]`,
// `?`, `|`, `->` - stays uncoloured.
//
// For the bare `list` and `fn` keywords (inside the parameterised
// list_type / fn_type rules), we recognise the first child token's
// literal text and emit on that. Tuple / enum forms inside
// list_type ("[T1, T2]", "[\"a\", \"b\"]") have no keyword and
// we skip those.
func collectTypeTokens(root *ts.Node, out *[]rawToken) {
	if root == nil {
		return
	}
	walkCSTNodes(root, func(n *ts.Node) {
		kind := n.Kind()
		if length, ok := typeKeywordLengths[kind]; ok {
			emitTypeToken(n, length, out)
			return
		}
		switch kind {
		case "list_type":
			// list_type covers `list` (the open-list keyword) and
			// the bracketed tuple / enum forms. Only the bare-
			// `list` shape has a keyword to paint - detect via
			// the literal first 4 bytes.
			if n.EndByte()-n.StartByte() >= 4 {
				emitTypeToken(n, 4, out)
			}
		case "fn_type":
			// fn_type starts with the literal `fn` token. Emit
			// just the first 2 bytes.
			if n.EndByte()-n.StartByte() >= 2 {
				emitTypeToken(n, 2, out)
			}
		}
	})
}

func emitTypeToken(n *ts.Node, length int, out *[]rawToken) {
	start := n.StartPosition()
	span := int(n.EndByte() - n.StartByte())
	if length < 0 || length > span {
		length = span
	}
	if length == 0 {
		return
	}
	*out = append(*out, rawToken{
		line:   int(start.Row),
		col:    int(start.Column),
		length: length,
		ttype:  TokenTypeType,
	})
}

func walkCSTNodes(n *ts.Node, visit func(*ts.Node)) {
	if n == nil {
		return
	}
	visit(n)
	count := n.ChildCount()
	for i := uint(0); i < count; i++ {
		child := n.Child(i)
		walkCSTNodes(child, visit)
	}
}

// tokenFromSpan packages a span into the rawToken shape. Length
// is byte-distance; the encoder translates to the negotiated
// character encoding later.
func tokenFromSpan(s rl.Span, ttype SemanticTokenType) rawToken {
	return rawToken{
		line:   s.StartRow,
		col:    s.StartCol,
		length: s.EndByte - s.StartByte,
		ttype:  ttype,
	}
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
	case check.SymParam, check.SymArg, check.SymCmdArg:
		// Args / cmd-args are declared parameters of the script
		// (the script's caller fills them in via CLI flags), so
		// they belong with fn params under the `parameter` token
		// type rather than the generic `variable`.
		return TokenTypeParameter, true
	case check.SymLocal, check.SymLoopVar, check.SymWith:
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
