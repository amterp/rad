package analysis

import (
	"rls/log"
	"rls/lsp"

	"github.com/amterp/rts"
	ts "github.com/tree-sitter/go-tree-sitter"
)

func (s *State) resolveDiagnostics(tree *rts.RslTree) []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)
	s.addInvalidNodes(&diagnostics, tree)
	s.addUnknownFunctions(&diagnostics, tree)
	return diagnostics
}

func (s *State) addInvalidNodes(diagnostics *[]lsp.Diagnostic, tree *rts.RslTree) {
	invalidNodes := tree.FindInvalidNodes()

	if len(invalidNodes) > 0 {
		log.L.Infof("Found %d invalid nodes", len(invalidNodes))
	}

	for _, node := range invalidNodes {
		rang := lsp.NewRangeFromTsNode(node)
		*diagnostics = append(*diagnostics, lsp.NewDiagnostic(rang, lsp.Err, "RSL Language Server", "Invalid node"))
	}
}

// todo this needs to be updated since lambdas/functions have been added
func (s *State) addUnknownFunctions(diagnostics *[]lsp.Diagnostic, tree *rts.RslTree) {
	calls := tree.FindCalls()

	unknownFuncNameNodes := make([]*ts.Node, 0)
	for _, call := range calls {
		if !s.rslFunctions.Contains(call.Name) {
			unknownFuncNameNodes = append(unknownFuncNameNodes, call.NameNode)
		}
	}

	if len(unknownFuncNameNodes) > 0 {
		log.L.Infof("Found %d unknown calls", len(unknownFuncNameNodes))
	}

	for _, node := range unknownFuncNameNodes {
		rang := lsp.NewRangeFromTsNode(node)
		*diagnostics = append(*diagnostics, lsp.NewDiagnostic(rang, lsp.Err, "RSL Language Server", "Unknown function"))
	}
}
