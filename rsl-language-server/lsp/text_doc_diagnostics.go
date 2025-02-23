package lsp

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
