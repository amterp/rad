package analysis

import (
	"rls/log"
	"rls/lsp"
)

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
	shebang, err := doc.tree.FindShebang()
	log.L.Infow("Searched for shebang", "err", err, "shebang", shebang)
	if shebang == nil || shebang.StartPos().Row != 0 {
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
