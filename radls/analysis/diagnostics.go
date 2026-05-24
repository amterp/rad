package analysis

import (
	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
)

func (s *State) resolveDiagnostics(doc *DocState) []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)

	result, err := doc.checker.Check()
	if err == nil {
		s.addCheckerDiagnotics(&diagnostics, result, doc)
	} else {
		log.L.Errorf("Failed to check script: %v", err)
	}
	return diagnostics
}

func (s *State) addCheckerDiagnotics(diagnostics *[]lsp.Diagnostic, checkResult check.Result, doc *DocState) {
	checkDiagnostics := checkResult.Diagnostics

	log.L.Infof("Found %d checker diagnostics", len(checkDiagnostics))

	for _, checkD := range checkDiagnostics {
		lspRange := s.toLspRange(checkD.Range, doc)
		*diagnostics = append(*diagnostics, lsp.NewDiagnosticFromCheckWithRange(checkD, lspRange))
	}
}

// toLspRange converts a check.Range (utf-8 byte columns, the tree-sitter
// native) into an LSP Range in the encoding the client negotiated.
func (s *State) toLspRange(r check.Range, doc *DocState) lsp.Range {
	idx := doc.lineIndex
	enc := s.encoding
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
