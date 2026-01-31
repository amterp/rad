package lsp

import (
	"github.com/amterp/rad/core"
	"github.com/amterp/rad/rts/check"
)

type DiagnosticSeverity int

const (
	Err  DiagnosticSeverity = 1
	Warn DiagnosticSeverity = 2
	Info DiagnosticSeverity = 3
	Hint DiagnosticSeverity = 4
)

type Diagnostic struct {
	Range    Range              `json:"range"`
	Severity DiagnosticSeverity `json:"severity"`
	/**
	 * A human-readable string describing the source of this
	 * diagnostic, e.g. 'typescript' or 'super lint'.
	 */
	Source string `json:"source"`
	/** The diagnostic's message. */
	Message string `json:"message"`
}

func NewDiagnostic(rang Range, severity DiagnosticSeverity, source, msg string) Diagnostic {
	return Diagnostic{
		Range:    rang,
		Severity: severity,
		Source:   source,
		Message:  msg,
	}
}

func NewDiagnosticFromCheck(checkD check.Diagnostic) Diagnostic {
	rang := NewRangeFromCheckNode(checkD.Range)
	var severity DiagnosticSeverity
	switch checkD.Severity {
	case check.Error:
		severity = Err
	case check.Hint:
		severity = Hint
	case check.Warning:
		severity = Warn
	case check.Info:
		severity = Info
	}
	return Diagnostic{
		Range:    rang,
		Severity: severity,
		Source:   "Rad Language Server",
		Message:  checkD.Message,
	}
}

// NewDiagnosticFromCore converts a core.Diagnostic to an LSP Diagnostic.
// This uses the primary span from the core.Diagnostic for the range.
func NewDiagnosticFromCore(coreD core.Diagnostic) Diagnostic {
	var rang Range
	if span := coreD.PrimarySpan(); span != nil {
		rang = NewRange(span.StartRow, span.StartCol, span.EndRow, span.EndCol)
	}

	var severity DiagnosticSeverity
	switch coreD.Severity {
	case core.SeverityError:
		severity = Err
	case core.SeverityWarning:
		severity = Warn
	case core.SeverityNote:
		severity = Hint
	default:
		severity = Err
	}

	return Diagnostic{
		Range:    rang,
		Severity: severity,
		Source:   "Rad Language Server",
		Message:  coreD.Message,
	}
}

type PublishDiagnosticsParams struct {
	Uri         string       `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

func NewPublishDiagnosticsParams(uri string, diagnostics []Diagnostic) PublishDiagnosticsParams {
	return PublishDiagnosticsParams{
		Uri:         uri,
		Diagnostics: diagnostics,
	}
}
