package lsp

import (
	"github.com/amterp/rad/radls/com"
)

// InitializeParams is a partial view of the LSP `initialize` request. We
// decode only the fields we actually use; the rest of the spec is large
// and most of it isn't yet load-bearing for radls.
type InitializeParams struct {
	ClientInfo   *ClientInfo        `json:"clientInfo"`
	Capabilities ClientCapabilities `json:"capabilities"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// ClientCapabilities mirrors the subset of LSP 3.17's
// ClientCapabilities.general we care about. `PositionEncodings` is the
// client's offered list - we negotiate one and echo our pick in
// ServerCapabilities.PositionEncoding.
type ClientCapabilities struct {
	General *GeneralClientCapabilities `json:"general,omitempty"`
}

type GeneralClientCapabilities struct {
	PositionEncodings []string `json:"positionEncodings,omitempty"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

// ServerCapabilities is what the server reports back at initialize.
// PositionEncoding is the encoding we picked from the client's offered
// list; clients use it for every subsequent request/response that
// contains line/character positions.
type ServerCapabilities struct {
	TextDocumentSync       int32                   `json:"textDocumentSync"`
	HoverProvider          bool                    `json:"hoverProvider"`
	DefinitionProvider     bool                    `json:"definitionProvider"`
	DocumentSymbolProvider bool                    `json:"documentSymbolProvider"`
	ReferencesProvider     bool                    `json:"referencesProvider"`
	CodeActionProvider     bool                    `json:"codeActionProvider"`
	RenameProvider         bool                    `json:"renameProvider"`
	CompletionProvider     map[string]any          `json:"completionProvider"`
	SemanticTokensProvider *SemanticTokensProvider `json:"semanticTokensProvider,omitempty"`
	PositionEncoding       string                  `json:"positionEncoding,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// NewInitializeResult builds the server's response to `initialize`.
// The legend is passed in (rather than fixed here) because the
// token-type vocabulary belongs to the analysis layer - keeping
// the lsp package free of analysis-internal knowledge avoids an
// import cycle and keeps the wire-types file scope-blind.
func NewInitializeResult(positionEncoding string, semLegend SemanticTokensLegend) InitializeResult {
	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync:       1,
			HoverProvider:          true,
			DefinitionProvider:     true,
			DocumentSymbolProvider: true,
			ReferencesProvider:     true,
			CodeActionProvider:     true,
			RenameProvider:         true,
			CompletionProvider: map[string]any{
				"triggerCharacters": []string{".", "#", "$", "!", "/"},
			},
			SemanticTokensProvider: &SemanticTokensProvider{
				Legend: semLegend,
				Full:   true,
			},
			PositionEncoding: positionEncoding,
		},
		ServerInfo: ServerInfo{
			Name:    "radls",
			Version: com.RadlsVersion,
		},
	}
}
