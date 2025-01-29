package analysis

import (
	"rls/log"
	"rls/lsp"
	"strings"
)

type DocState struct {
	uri  string
	text string
}

func NewDocState(uri, text string) DocState {
	return DocState{uri: uri, text: text}
}

func (d *DocState) GetLine(line int) string {
	// todo wasteful implementation
	lines := strings.Split(d.text, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	return lines[line]
}

type State struct {
	// URI -> Text
	docs map[string]DocState
}

func NewState() *State {
	return &State{docs: make(map[string]DocState)}
}

func (s *State) AddDoc(uri, text string) {
	log.L.Infof("Adding doc %s", uri)
	s.docs[uri] = NewDocState(uri, text)
}

func (s *State) UpdateDoc(uri string, changes []lsp.TextDocumentContentChangeEvent) {
	doc := s.docs[uri]
	for _, change := range changes {
		doc.text = change.Text
	}
}

func (s *State) Complete(uri string, pos lsp.Pos) (result []lsp.CompletionItem, err error) {
	doc, ok := s.docs[uri]
	if !ok {
		return nil, nil // todo return error?
	}

	var items []lsp.CompletionItem
	addShebangCompletion(&items, doc, pos)
	return items, nil
}

func addShebangCompletion(i *[]lsp.CompletionItem, doc DocState, pos lsp.Pos) {
	// todo use tree sitter to check for shebang node?

	if pos.Line != 0 {
		return
	}

	//line := doc.GetLine(pos.Line)

	*i = append(*i, lsp.CompletionItem{
		Label:  RadShebang,
		Detail: "Shebang for rad",
		//TextEdit: lsp.NewTextEdit(lsp.NewLineRange(0, 0, len(line)), RadShebang),
	})
}
