package lsp

import (
	"rsl-lsp/com"
)

type InitializeParams struct {
	ClientInfo *ClientInfo `json:"clientInfo"`
	// lots more in here
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	Capabilities ServerCapabilities `json:"capabilities"`
	ServerInfo   ServerInfo         `json:"serverInfo"`
}

type ServerCapabilities struct {
	TextDocumentSync int32 `json:"textDocumentSync"`
	//HoverProvider      bool           `json:"hoverProvider"`
	//DefinitionProvider bool           `json:"definitionProvider"`
	//CodeActionProvider bool           `json:"codeActionProvider"`
	//CompletionProvider map[string]any `json:"completionProvider"` // bool not supported by spec
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewInitializeResult() InitializeResult {
	return InitializeResult{
		Capabilities: ServerCapabilities{
			TextDocumentSync: 1,
			//HoverProvider:      true,
			//DefinitionProvider: true,
			//CodeActionProvider: true,
			//CompletionProvider: map[string]any{},
		},
		ServerInfo: ServerInfo{
			Name:    "RSL LSP",
			Version: com.RslLspVersion,
		},
	}
}
