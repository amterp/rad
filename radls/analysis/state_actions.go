package analysis

import (
	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"
)

// Complete runs the completion logic against a fixed document
// snapshot. Callers (server handlers) grab the snapshot themselves
// and pass it explicitly: the snapshot pin-points which version of
// the document this completion is meant for, even if the user types
// more characters before the response is sent.
//
// The result is the union of the shebang stub (only meaningful on
// line 0) and a scope-aware identifier list built from the AST
// and the resolved/type indexes. The shebang stays first so users
// who started a new file still get the "Add #!" suggestion at the
// top of the list.
func (s *State) Complete(snap *DocumentVersion, pos lsp.Pos) (result []lsp.CompletionItem, err error) {
	if snap == nil {
		return nil, nil
	}

	// Translate the incoming position from the client's encoding into a
	// utf-8 byte column so the rest of the analyzer can stay in
	// tree-sitter's native coordinate system.
	bytePos := toBytePos(pos, snap)

	items := make([]lsp.CompletionItem, 0)
	addShebangCompletion(&items, snap, bytePos)
	buildCompletions(&items, snap, bytePos)
	return items, nil
}

// CodeAction returns the available code actions for the given range
// against a fixed document snapshot. Same snapshot discipline as
// Complete: the caller's responsibility to grab a fresh one.
//
// Two sources of actions today:
//
//  1. The always-available shebang insertion, when the file is
//     missing one - this is whole-document, not range-scoped.
//  2. Diagnostic quick fixes - per-diagnostic actions for issues
//     whose range overlaps the request range. Structured fixes
//     (machine-applicable edits) come out as quickfix-kinded
//     CodeActions; suggestion-only diagnostics produce refactor-
//     kinded info actions with no edit.
func (s *State) CodeAction(snap *DocumentVersion, r lsp.Range) (result []lsp.CodeAction, err error) {
	if snap == nil {
		return nil, nil
	}

	actions := make([]lsp.CodeAction, 0)
	addShebangInsertion(&actions, snap)
	addDiagnosticQuickFixes(&actions, snap, toByteRange(r, snap))

	return actions, nil
}

// toBytePos converts an incoming LSP position from the client's encoding
// into a utf-8 byte column on the given snapshot. The line number passes
// through unchanged - LSP lines and our internal lines both count \n.
//
// The encoding comes from the snapshot, not the State, so the conversion
// stays consistent with whichever encoding was in effect when this
// version was built. If a session somehow renegotiates encoding (LSP
// 3.17 doesn't allow this but be defensive), older snapshots remain
// internally consistent.
func toBytePos(pos lsp.Pos, snap *DocumentVersion) lsp.Pos {
	return lsp.Pos{
		Line:      pos.Line,
		Character: snap.lineIndex.ColumnToByte(pos.Line, pos.Character, snap.encoding),
	}
}

// toByteRange is the Range-shaped counterpart of toBytePos.
func toByteRange(r lsp.Range, snap *DocumentVersion) lsp.Range {
	return lsp.Range{
		Start: toBytePos(r.Start, snap),
		End:   toBytePos(r.End, snap),
	}
}

// fromByteRange converts a Range expressed in utf-8 byte columns into
// the snapshot's encoding. Used when we construct a WorkspaceEdit
// from internal positions (e.g. tree-sitter node spans).
func fromByteRange(r lsp.Range, snap *DocumentVersion) lsp.Range {
	idx := snap.lineIndex
	enc := snap.encoding
	return lsp.Range{
		Start: lsp.Pos{
			Line:      r.Start.Line,
			Character: idx.ByteColumnTo(r.Start.Line, r.Start.Character, enc),
		},
		End: lsp.Pos{
			Line:      r.End.Line,
			Character: idx.ByteColumnTo(r.End.Line, r.End.Character, enc),
		},
	}
}

func addShebangInsertion(i *[]lsp.CodeAction, snap *DocumentVersion) {
	shebang, found := snap.tree.FindShebang()
	log.L.Infow("Searched for shebang", "found", found, "shebang", shebang)
	if shebang == nil || shebang.StartPos().Row != 0 {
		firstLine := snap.GetLine(0)
		log.L.Infow("First line does not have #!, adding insertion action", "line", firstLine)
		edit := lsp.NewWorkspaceEdit()
		// The insertion site is (0,0,0,0) - identical in every encoding -
		// but we route it through fromByteRange anyway so future code
		// actions inherit the right conversion habit.
		insertRange := fromByteRange(lsp.NewLineRange(0, 0, 0), snap)
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
