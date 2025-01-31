package analysis

import (
	"rls/log"
	"rls/lsp"
	"strings"

	"github.com/amterp/rts"
)

type DocState struct {
	uri  string
	text string
	tree *rts.RtsTree
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
	rts *rts.RslTreeSitter
	// URI -> Text
	docs map[string]*DocState
}

func NewState() *State {
	rslTs, err := rts.NewRts()
	if err != nil {
		log.L.Fatalw("Failed to create RSL tree sitter", "err", err)
	}

	return &State{
		rts:  rslTs,
		docs: make(map[string]*DocState),
	}
}

func (s *State) NewDocState(uri, text string) *DocState {
	tree, err := s.rts.Parse(text)
	if err != nil {
		log.L.Errorw("Failed to parse doc", "uri", uri, "err", err)
		return nil // todo putting nil into the map??
	}
	return &DocState{
		uri:  uri,
		text: text,
		tree: tree,
	}
}

func (s *State) AddDoc(uri, text string) {
	log.L.Infof("Adding doc %s", uri)
	s.docs[uri] = s.NewDocState(uri, text)
}

func (s *State) UpdateDoc(uri string, changes []lsp.TextDocumentContentChangeEvent) {
	doc := s.docs[uri]
	for _, change := range changes {
		log.L.Infow("Updating doc", "uri", uri)
		log.L.Debugf("Tree before: %s", doc.tree.String())
		doc.tree.Update(change.Text)
		log.L.Debugf("Tree after: %s", doc.tree.String())
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

func (s *State) CodeAction(uri string, r lsp.Range) (result []lsp.CodeAction, err error) {
	doc, ok := s.docs[uri]
	if !ok {
		return nil, nil // todo return error?
	}

	var actions []lsp.CodeAction
	addShebangInsertion(&actions, doc)

	return actions, nil
}

func addShebangInsertion(i *[]lsp.CodeAction, doc *DocState) {
	shebang, ok := doc.tree.GetShebang()
	log.L.Infow("Searched for shebang", "ok", ok, "shebang", shebang)
	if !ok || shebang.StartPos.Row != 0 {
		firstLine := doc.GetLine(0)
		log.L.Infow("First line does not have #!, adding insertion action", "line", firstLine)
		edit := lsp.NewWorkspaceEdit()
		edit.AddEdit(doc.uri, lsp.NewLineRange(0, 0, 0), RadShebang+"\n")
		action := lsp.NewCodeActionEdit("Add shebang", edit)
		*i = append(*i, action)
	}
}

func addShebangCompletion(i *[]lsp.CompletionItem, doc *DocState, pos lsp.Pos) {
	// todo use tree sitter to check for shebang node?

	if pos.Line != 0 {
		return
	}

	//line := doc.GetLine(pos.Line)

	*i = append(*i, lsp.CompletionItem{
		Label:  RadShebang,
		Detail: "Shebang for rad",
		// todo add docs
		//TextEdit: lsp.NewTextEdit(lsp.NewLineRange(0, 0, len(line)), RadShebang),
	})
}
