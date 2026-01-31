package core

import (
	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// Severity indicates the severity level of a diagnostic.
type Severity int

const (
	SeverityError Severity = iota
	SeverityWarning
	SeverityNote
)

func (s Severity) String() string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warning"
	case SeverityNote:
		return "note"
	default:
		return "unknown"
	}
}

// Span represents a location range in source code.
// All positions are 0-indexed.
type Span struct {
	File      string
	StartByte int
	EndByte   int
	StartRow  int
	StartCol  int
	EndRow    int
	EndCol    int
}

// NewSpanFromNode creates a Span from a tree-sitter node and file path.
func NewSpanFromNode(node *ts.Node, file string) Span {
	return Span{
		File:      file,
		StartByte: int(node.StartByte()),
		EndByte:   int(node.EndByte()),
		StartRow:  int(node.StartPosition().Row),
		StartCol:  int(node.StartPosition().Column),
		EndRow:    int(node.EndPosition().Row),
		EndCol:    int(node.EndPosition().Column),
	}
}

// StartLine returns the 1-indexed start line number for display.
func (s Span) StartLine() int {
	return s.StartRow + 1
}

// StartColumn returns the 1-indexed start column number for display.
func (s Span) StartColumn() int {
	return s.StartCol + 1
}

// Label represents a labeled span in a diagnostic.
type Label struct {
	Span    Span
	Message string
	Primary bool // true = ^^^^ (red), false = ---- (blue)
}

// NewPrimaryLabel creates a primary label (shown with ^^^^ in red).
func NewPrimaryLabel(span Span, message string) Label {
	return Label{
		Span:    span,
		Message: message,
		Primary: true,
	}
}

// NewSecondaryLabel creates a secondary label (shown with ---- in blue).
func NewSecondaryLabel(span Span, message string) Label {
	return Label{
		Span:    span,
		Message: message,
		Primary: false,
	}
}

// Diagnostic represents a single diagnostic message with optional multi-span context.
type Diagnostic struct {
	Severity Severity
	Code     rl.Error  // From rts/rl/errors.go
	Message  string    // One-line summary
	Labels   []Label   // Primary + secondary spans
	Hints    []string  // "= help: ..." lines
	Source   string    // Complete source for rendering
}

// NewDiagnostic creates a diagnostic with a single primary label.
func NewDiagnostic(severity Severity, code rl.Error, message string, source string, primarySpan Span) Diagnostic {
	return Diagnostic{
		Severity: severity,
		Code:     code,
		Message:  message,
		Labels:   []Label{NewPrimaryLabel(primarySpan, "")},
		Source:   source,
	}
}

// NewDiagnosticWithLabels creates a diagnostic with multiple labels.
func NewDiagnosticWithLabels(severity Severity, code rl.Error, message string, source string, labels []Label) Diagnostic {
	return Diagnostic{
		Severity: severity,
		Code:     code,
		Message:  message,
		Labels:   labels,
		Source:   source,
	}
}

// WithHint adds a hint to the diagnostic and returns the modified diagnostic.
func (d Diagnostic) WithHint(hint string) Diagnostic {
	d.Hints = append(d.Hints, hint)
	return d
}

// WithHints adds multiple hints to the diagnostic and returns the modified diagnostic.
func (d Diagnostic) WithHints(hints ...string) Diagnostic {
	d.Hints = append(d.Hints, hints...)
	return d
}

// WithSecondaryLabel adds a secondary label to the diagnostic and returns the modified diagnostic.
func (d Diagnostic) WithSecondaryLabel(span Span, message string) Diagnostic {
	d.Labels = append(d.Labels, NewSecondaryLabel(span, message))
	return d
}

// PrimarySpan returns the first primary span, or nil if none exists.
func (d Diagnostic) PrimarySpan() *Span {
	for _, label := range d.Labels {
		if label.Primary {
			return &label.Span
		}
	}
	return nil
}

// DiagnosticCollector accumulates diagnostics up to a configurable limit.
// Tree-sitter's error recovery can cause cascades where one real error spawns many,
// so limiting prevents wall-of-text noise while still showing patterns.
type DiagnosticCollector struct {
	diagnostics  []Diagnostic
	limit        int
	totalEmitted int // track total for "...and X more" message
}

// DefaultDiagnosticLimit is the default maximum number of diagnostics to collect.
const DefaultDiagnosticLimit = 10

// NewDiagnosticCollector creates a collector with the default limit.
func NewDiagnosticCollector() *DiagnosticCollector {
	return &DiagnosticCollector{
		diagnostics: make([]Diagnostic, 0),
		limit:       DefaultDiagnosticLimit,
	}
}

// NewDiagnosticCollectorWithLimit creates a collector with a custom limit.
func NewDiagnosticCollectorWithLimit(limit int) *DiagnosticCollector {
	return &DiagnosticCollector{
		diagnostics: make([]Diagnostic, 0),
		limit:       limit,
	}
}

// Add adds a diagnostic to the collector.
// Returns false when the limit is reached, signaling that the caller should stop
// producing diagnostics.
func (c *DiagnosticCollector) Add(d Diagnostic) bool {
	c.totalEmitted++
	if len(c.diagnostics) >= c.limit {
		return false
	}
	c.diagnostics = append(c.diagnostics, d)
	return true
}

// Diagnostics returns all collected diagnostics.
func (c *DiagnosticCollector) Diagnostics() []Diagnostic {
	return c.diagnostics
}

// Count returns the number of diagnostics collected.
func (c *DiagnosticCollector) Count() int {
	return len(c.diagnostics)
}

// TotalEmitted returns the total number of diagnostics that were attempted to be added,
// including those beyond the limit.
func (c *DiagnosticCollector) TotalEmitted() int {
	return c.totalEmitted
}

// Remaining returns how many more diagnostics were emitted beyond the limit.
func (c *DiagnosticCollector) Remaining() int {
	if c.totalEmitted > c.limit {
		return c.totalEmitted - c.limit
	}
	return 0
}

// HasErrors returns true if there are any error-severity diagnostics.
func (c *DiagnosticCollector) HasErrors() bool {
	for _, d := range c.diagnostics {
		if d.Severity == SeverityError {
			return true
		}
	}
	return false
}

// IsEmpty returns true if no diagnostics have been collected.
func (c *DiagnosticCollector) IsEmpty() bool {
	return len(c.diagnostics) == 0
}

// AtLimit returns true if the collector has reached its limit.
func (c *DiagnosticCollector) AtLimit() bool {
	return len(c.diagnostics) >= c.limit
}

// --- Conversion from check.Diagnostic ---

// convertCheckSeverity converts check.Severity to core.Severity.
func convertCheckSeverity(s check.Severity) Severity {
	switch s {
	case check.Error:
		return SeverityError
	case check.Warning:
		return SeverityWarning
	case check.Hint, check.Info:
		return SeverityNote
	default:
		return SeverityError
	}
}

// NewDiagnosticFromCheck converts a check.Diagnostic to a core.Diagnostic.
// The file parameter is used for the span's file path.
func NewDiagnosticFromCheck(d check.Diagnostic, file string) Diagnostic {
	// Create span from check.Range
	span := Span{
		File:     file,
		StartRow: d.Range.Start.Line,
		StartCol: d.Range.Start.Character,
		EndRow:   d.Range.End.Line,
		EndCol:   d.Range.End.Character,
		// Byte offsets aren't available in check.Range, but we can compute them
		// from the source if needed. For now, leave at 0 - the renderer uses
		// row/col for display anyway.
		StartByte: 0,
		EndByte:   0,
	}

	// Determine the error code
	var code rl.Error
	if d.Code != nil {
		code = *d.Code
	} else {
		code = rl.ErrGenericRuntime
	}

	diag := Diagnostic{
		Severity: convertCheckSeverity(d.Severity),
		Code:     code,
		Message:  d.Message,
		Labels:   []Label{NewPrimaryLabel(span, "")},
		Source:   d.OriginalSrc,
	}

	// Add suggestion as a hint if present
	if d.Suggestion != nil && *d.Suggestion != "" {
		diag.Hints = append(diag.Hints, *d.Suggestion)
	}

	return diag
}
