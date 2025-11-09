package analysis

import (
	"github.com/amterp/rad/lsp-server/log"
	"github.com/amterp/rad/lsp-server/lsp"

	"github.com/amterp/rad/rts/check"
)

func (s *State) resolveDiagnostics(checker check.RadChecker) []lsp.Diagnostic {
	diagnostics := make([]lsp.Diagnostic, 0)

	result, err := checker.CheckDefault()
	if err == nil {
		s.addCheckerDiagnotics(&diagnostics, result)
	} else {
		log.L.Errorf("Failed to check script: %v", err)
	}
	return diagnostics
}

func (s *State) addCheckerDiagnotics(diagnostics *[]lsp.Diagnostic, checkResult check.Result) {
	checkDiagnostics := checkResult.Diagnostics

	log.L.Infof("Found %d checker diagnostics", len(checkDiagnostics))

	for _, checkD := range checkDiagnostics {
		*diagnostics = append(*diagnostics, lsp.NewDiagnosticFromCheck(checkD))
	}
}
