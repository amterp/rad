package lsp

type TextDocumentItem struct {
	/**
	 * The text document's URI.
	 */
	Uri string `json:"uri"`
	/**
	 * The text document's language identifier.
	 */
	LanguageId string `json:"languageId"`
	/**
	 * The version number of this document (it will increase after each
	 * change, including undo/redo).
	 */
	Version int `json:"version"`
	/**
	 * The content of the opened text document.
	 */
	Text string `json:"text"`
}

// https://microsoft.github.io/language-server-protocol/specifications/lsp/3.17/specification/#didOpenTextDocumentParams
type DidOpenTextDocumentParams struct {
	/**
	 * The document that was opened.
	 */
	TextDocument TextDocumentItem `json:"textDocument"`
}

type VersionedTextDocumentIdentifier struct {
	TextDocumentIdentifier
	Version int `json:"version"`
}

/**
 * An event describing a change to a text document. If only a text is provided
 * it is considered to be the full content of the document.
 */
type TextDocumentContentChangeEvent struct {
	/**
	 * The new text of the whole document.
	 */
	Text string `json:"text"`
}

type DidChangeTextDocumentParams struct {
	/**
	 * The document that did change. The version number points
	 * to the version after all provided content changes have
	 * been applied.
	 */
	TextDocument VersionedTextDocumentIdentifier `json:"textDocument"`

	/**
	 * The actual content changes. The content changes describe single state
	 * changes to the document. So if there are two content changes c1 (at
	 * array index 0) and c2 (at array index 1) for a document in state S then
	 * c1 moves the document from S to S' and c2 from S' to S''. So c1 is
	 * computed on the state S and c2 is computed on the state S'.
	 *
	 * To mirror the content of a document using change events use the following
	 * approach:
	 * - start with the same initial content
	 * - apply the 'textDocument/didChange' notifications in the order you
	 *   receive them.
	 * - apply the `TextDocumentContentChangeEvent`s in a single notification
	 *   in the order you receive them.
	 */
	ContentChanges []TextDocumentContentChangeEvent `json:"contentChanges"`
}

type Pos struct {
	/** Zero-indexed */
	Line int `json:"line"`
	/** Zero-indexed */
	Character int `json:"character"`
}

func NewPos(line int, character int) Pos {
	return Pos{
		Line:      line,
		Character: character,
	}
}

type Range struct {
	Start Pos `json:"start"`
	End   Pos `json:"end"`
}

func NewRange(startLine, starChar, endLine, endChar int) Range {
	return Range{
		Start: NewPos(startLine, starChar),
		End:   NewPos(endLine, endChar),
	}
}

func NewLineRange(line, start, end int) Range {
	return NewRange(line, start, line, end)
}

type Location struct {
	Uri   string `json:"uri"`
	Range Range  `json:"range"`
}

type TextDocumentIdentifier struct {
	Uri string `json:"uri"`
}

type TextDocumentPositionParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	Position     Pos                    `json:"position"`
}

type CompletionParams struct {
	TextDocumentPositionParams
}

// HoverParams is the textDocument/hover request payload. Same shape
// as completion - just a position - but we keep distinct types so
// any spec-level divergence (work-done tokens, etc.) lands cleanly.
type HoverParams struct {
	TextDocumentPositionParams
}

// DefinitionParams is the textDocument/definition request payload.
// The response can be a single Location, an array of Locations, or
// null - we return either nil (encoded as null) or a single
// Location since Rad has exactly one decl site per symbol today.
type DefinitionParams struct {
	TextDocumentPositionParams
}

// DocumentSymbolParams is the textDocument/documentSymbol request
// payload. The whole document is implied; no position or range is
// passed because the response is "outline of the file."
type DocumentSymbolParams struct {
	TextDocument TextDocumentIdentifier `json:"textDocument"`
}

// SymbolKind matches the LSP 3.17 SymbolKind enum. We only use a
// small slice today; the full list is in the spec but adding
// constants we don't emit just clutters the file.
type SymbolKind int

const (
	SymbolKindModule    SymbolKind = 2
	SymbolKindFunction  SymbolKind = 12
	SymbolKindVariable  SymbolKind = 13
	SymbolKindNamespace SymbolKind = 3
)

// DocumentSymbol is the hierarchical outline shape (preferred over
// the legacy flat SymbolInformation list). `Range` is the symbol's
// whole-extent span; `SelectionRange` is the name token specifically
// (what the editor highlights). The two often differ - a function
// `Range` covers the whole body, while `SelectionRange` covers just
// the name - so we always set both even when they're equal.
type DocumentSymbol struct {
	Name           string           `json:"name"`
	Detail         string           `json:"detail,omitempty"`
	Kind           SymbolKind       `json:"kind"`
	Range          Range            `json:"range"`
	SelectionRange Range            `json:"selectionRange"`
	Children       []DocumentSymbol `json:"children,omitempty"`
}

// MarkupKind matches the LSP 3.17 MarkupKind enum. We always reply
// in markdown today; plaintext is a future option if a client
// declines markdown.
type MarkupKind string

const (
	MarkupPlainText MarkupKind = "plaintext"
	MarkupMarkdown  MarkupKind = "markdown"
)

// MarkupContent is the rich-text payload format for hover and a few
// other LSP responses. We emit markdown so we can render fenced code
// blocks (`rad` language identifier) which clients style as syntax.
type MarkupContent struct {
	Kind  MarkupKind `json:"kind"`
	Value string     `json:"value"`
}

// Hover is the response to textDocument/hover. Range is optional but
// providing it lets the editor highlight the token the hover applies
// to, which feels much better than a bare popup.
type Hover struct {
	Contents MarkupContent `json:"contents"`
	Range    *Range        `json:"range,omitempty"`
}

type TextEdit struct {
	Range   Range  `json:"range"`
	NewText string `json:"newText"`
}

func NewTextEdit(rang Range, newText string) *TextEdit {
	return &TextEdit{
		Range:   rang,
		NewText: newText,
	}
}

type CompletionItem struct {
	/**
	 * The label of this completion item.
	 *
	 * The label property is also by default the text that
	 * is inserted when selecting this completion, UNLESS TEXTEDIT PROVIDED.
	 *
	 * If label details are provided the label itself should
	 * be an unqualified name of the completion item.
	 */
	Label    string    `json:"label"`
	Detail   string    `json:"detail"`
	Doc      string    `json:"documentation,omitempty"`
	TextEdit *TextEdit `json:"textEdit,omitempty"`
	// insertText might be useful for inserting imports at the top, as needed
	//Kind todo for icon
}

func NewCompletionItem(label, detail, doc string) CompletionItem {
	return CompletionItem{
		Label:  label,
		Detail: detail,
		Doc:    doc,
	}
}

type CodeActionParams struct {
	// The document in which the command was invoked.
	TextDocument TextDocumentIdentifier `json:"textDocument"`
	// The range for which the command was invoked.
	Range Range `json:"range"`
}

type CodeAction struct {
	Title   string         `json:"title"`
	Edit    *WorkspaceEdit `json:"edit,omitempty"`
	Command *Command       `json:"command,omitempty"`
}

func NewCodeActionEdit(title string, edit WorkspaceEdit) CodeAction {
	return CodeAction{
		Title: title,
		Edit:  &edit,
	}
}

type WorkspaceEdit struct {
	// URI -> TextEdits
	Changes map[string][]TextEdit `json:"changes"`
}

func NewWorkspaceEdit() WorkspaceEdit {
	return WorkspaceEdit{
		Changes: make(map[string][]TextEdit),
	}
}

func (w *WorkspaceEdit) AddEdit(uri string, rang Range, text string) {
	w.Changes[uri] = append(w.Changes[uri], *NewTextEdit(rang, text))
}

type Command struct {
	Title     string        `json:"title"`
	Command   string        `json:"command"`
	Arguments []interface{} `json:"arguments,omitempty"`
}
