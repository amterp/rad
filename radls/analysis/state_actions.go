package analysis

import (
	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"
)

func (s *State) Complete(uri string, pos lsp.Pos) (result []lsp.CompletionItem, err error) {
	snap := s.Snapshot(uri)
	if snap == nil {
		return nil, nil // todo return error?
	}

	// Translate the incoming position from the client's encoding into a
	// utf-8 byte column so the rest of the analyzer can stay in
	// tree-sitter's native coordinate system. Today addShebangCompletion
	// only looks at the line number, but completion will grow.
	bytePos := s.toBytePos(pos, snap)

	var items []lsp.CompletionItem
	addShebangCompletion(&items, snap, bytePos)
	return items, nil
}

func (s *State) CodeAction(uri string, r lsp.Range) (result []lsp.CodeAction, err error) {
	snap := s.Snapshot(uri)
	if snap == nil {
		return nil, nil // todo return error?
	}

	// We don't yet need the range for picking code actions (the only
	// action is shebang insertion, which is whole-document), but we
	// translate to byte coords for future code actions that will care.
	_ = s.toByteRange(r, snap)

	var actions []lsp.CodeAction
	addShebangInsertion(&actions, snap, s)

	return actions, nil
}

// toBytePos converts an incoming LSP position from the client's encoding
// into a utf-8 byte column on the given snapshot. The line number passes
// through unchanged - LSP lines and our internal lines both count \n.
func (s *State) toBytePos(pos lsp.Pos, snap *DocumentVersion) lsp.Pos {
	return lsp.Pos{
		Line:      pos.Line,
		Character: snap.lineIndex.ColumnToByte(pos.Line, pos.Character, s.encoding),
	}
}

// toByteRange is the Range-shaped counterpart of toBytePos.
func (s *State) toByteRange(r lsp.Range, snap *DocumentVersion) lsp.Range {
	return lsp.Range{
		Start: s.toBytePos(r.Start, snap),
		End:   s.toBytePos(r.End, snap),
	}
}

// fromByteRange converts a Range expressed in utf-8 byte columns into
// the client's negotiated encoding. Used when we construct a WorkspaceEdit
// from internal positions (e.g. tree-sitter node spans).
func (s *State) fromByteRange(r lsp.Range, snap *DocumentVersion) lsp.Range {
	idx := snap.lineIndex
	return lsp.Range{
		Start: lsp.Pos{
			Line:      r.Start.Line,
			Character: idx.ByteColumnTo(r.Start.Line, r.Start.Character, s.encoding),
		},
		End: lsp.Pos{
			Line:      r.End.Line,
			Character: idx.ByteColumnTo(r.End.Line, r.End.Character, s.encoding),
		},
	}
}

func addShebangInsertion(i *[]lsp.CodeAction, snap *DocumentVersion, s *State) {
	shebang, err := snap.tree.FindShebang()
	log.L.Infow("Searched for shebang", "err", err, "shebang", shebang)
	if shebang == nil || shebang.StartPos().Row != 0 {
		firstLine := snap.GetLine(0)
		log.L.Infow("First line does not have #!, adding insertion action", "line", firstLine)
		edit := lsp.NewWorkspaceEdit()
		// The insertion site is (0,0,0,0) - identical in every encoding -
		// but we route it through fromByteRange anyway so future code
		// actions inherit the right conversion habit.
		insertRange := s.fromByteRange(lsp.NewLineRange(0, 0, 0), snap)
		edit.AddEdit(snap.uri, insertRange, RadShebang+"\n")
		action := lsp.NewCodeActionEdit("Add shebang", edit)
		*i = append(*i, action)
	}
}

func addShebangCompletion(i *[]lsp.CompletionItem, snap *DocumentVersion, pos lsp.Pos) {
	// todo use tree sitter to check for shebang node?

	if pos.Line != 0 {
		return
	}

	*i = append(*i, lsp.CompletionItem{
		Label:  RadShebang,
		Detail: "Shebang for rad",
		// todo add docs
		//TextEdit: lsp.NewTextEdit(lsp.NewLineRange(0, 0, len(line)), RadShebang),
	})
}
