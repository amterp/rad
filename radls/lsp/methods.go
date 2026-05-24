package lsp

import "encoding/json"

const (
	INITIALIZE             = "initialize"
	CANCEL_REQUEST         = "$/cancelRequest"
	TD_DID_OPEN            = "textDocument/didOpen"
	TD_DID_CHANGE          = "textDocument/didChange"
	TD_COMPLETION          = "textDocument/completion"
	TD_CODE_ACTION         = "textDocument/codeAction"
	TD_HOVER               = "textDocument/hover"
	TD_DEFINITION          = "textDocument/definition"
	TD_DOCUMENT_SYMBOL     = "textDocument/documentSymbol"
	TD_REFERENCES          = "textDocument/references"
	TD_PUBLISH_DIAGNOSTICS = "textDocument/publishDiagnostics"
)

// CancelParams matches the LSP 3.17 $/cancelRequest payload. The
// `id` field may be a JSON number or string (the same as a request
// id), so we keep it as raw JSON and compare byte-for-byte.
type CancelParams struct {
	Id json.RawMessage `json:"id"`
}
