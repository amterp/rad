package analysis

import (
	"rls/log"
	"rls/lsp"

	"github.com/amterp/rad/rts/check"

	"github.com/amterp/rad/rts"
	ts "github.com/tree-sitter/go-tree-sitter"
)

func (s *State) resolveDiagnostics(tree *rts.RslTree, checker check.RadChecker) []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)
	result, err := checker.CheckDefault()
	if err == nil {
		s.addCheckerDiagnotics(&diagnostics, result)
	} else {
		log.L.Errorf("Failed to check script: %v", err)
	}
	s.addUnknownFunctions(&diagnostics, tree)
	return diagnostics
}

func (s *State) addCheckerDiagnotics(diagnostics *[]lsp.Diagnostic, checkResult check.Result) {
	checkDiagnostics := checkResult.Diagnostics

	log.L.Infof("Found %d checker diagnostics", len(checkDiagnostics))

	for _, checkD := range checkDiagnostics {
		*diagnostics = append(*diagnostics, lsp.NewDiagnosticFromCheck(checkD))
	}
}

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
		// todo this needs to be updated since lambdas/functions have been added. We just warn instead until we have a better script understanding.
		*diagnostics = append(*diagnostics, lsp.NewDiagnostic(rang, lsp.Warn, "RSL Language Server", "Non-builtin function"))
	}
}
