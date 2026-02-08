package check

import (
	"strings"

	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type Severity int

const (
	Hint Severity = iota
	Warning
	Info
	Error
)

func (s Severity) String() string {
	switch s {
	case Hint:
		return "Hint"
	case Info:
		return "Info"
	case Warning:
		return "Warn"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

// Diagnostic represents a static analysis diagnostic. This is the canonical
// diagnostic type for the check system; both lsp.Diagnostic and core.Diagnostic
// are derived from it. Severity levels align with the LSP protocol (4 levels);
// core collapses Hint/Info into Note for its 3-level Rust-style CLI output.
type Diagnostic struct {
	OriginalSrc string // Complete original src
	Range       Range
	RangedSrc   string // Src for just the Range
	LineSrc     string // Src for the line at the start of Range
	Severity    Severity
	Message     string
	Code        *rl.Error
	Suggestion  *string // Optional suggestion for fixing the error (rendered as "Try: ...")
}

func NewDiagnosticFromNode(
	node *ts.Node,
	originalSrc string,
	severity Severity,
	msg string,
	code *rl.Error,
) Diagnostic {
	line := int(node.StartPosition().Row)
	rang := Range{
		Start: Pos{
			Line:      line,
			Character: int(node.StartPosition().Column),
		},
		End: Pos{
			Line:      int(node.EndPosition().Row),
			Character: int(node.EndPosition().Column),
		},
	}
	lineSrc := strings.Split(originalSrc, "\n")[line]
	return Diagnostic{
		OriginalSrc: originalSrc,
		Range:       rang,
		RangedSrc:   originalSrc[node.StartByte():node.EndByte()],
		LineSrc:     lineSrc,
		Severity:    severity,
		Message:     msg,
		Code:        code,
	}
}

func NewDiagnosticError(node *ts.Node, originalSrc string, msg string, code rl.Error) Diagnostic {
	return NewDiagnosticFromNode(node, originalSrc, Error, msg, &code)
}

func NewDiagnosticErrorWithSuggestion(node *ts.Node, originalSrc string, msg string, code rl.Error, suggestion *string) Diagnostic {
	diag := NewDiagnosticFromNode(node, originalSrc, Error, msg, &code)
	diag.Suggestion = suggestion
	return diag
}

func NewDiagnosticHint(node *ts.Node, originalSrc string, msg string, code rl.Error) Diagnostic {
	return NewDiagnosticFromNode(node, originalSrc, Hint, msg, &code)
}

func NewDiagnosticWarn(node *ts.Node, originalSrc string, msg string, code rl.Error) Diagnostic {
	return NewDiagnosticFromNode(node, originalSrc, Warning, msg, &code)
}

// NewDiagnosticFromSpan creates a Diagnostic from an AST span instead of a CST node.
func NewDiagnosticFromSpan(
	span rl.Span,
	originalSrc string,
	severity Severity,
	msg string,
	code *rl.Error,
) Diagnostic {
	line := span.StartRow
	rang := Range{
		Start: Pos{
			Line:      line,
			Character: span.StartCol,
		},
		End: Pos{
			Line:      span.EndRow,
			Character: span.EndCol,
		},
	}
	lineSrc := strings.Split(originalSrc, "\n")[line]
	return Diagnostic{
		OriginalSrc: originalSrc,
		Range:       rang,
		RangedSrc:   originalSrc[span.StartByte:span.EndByte],
		LineSrc:     lineSrc,
		Severity:    severity,
		Message:     msg,
		Code:        code,
	}
}

func NewDiagnosticErrorFromSpan(span rl.Span, originalSrc string, msg string, code rl.Error) Diagnostic {
	return NewDiagnosticFromSpan(span, originalSrc, Error, msg, &code)
}

func NewDiagnosticHintFromSpan(span rl.Span, originalSrc string, msg string, code rl.Error) Diagnostic {
	return NewDiagnosticFromSpan(span, originalSrc, Hint, msg, &code)
}

func NewDiagnosticWarnFromSpan(span rl.Span, originalSrc string, msg string, code rl.Error) Diagnostic {
	return NewDiagnosticFromSpan(span, originalSrc, Warning, msg, &code)
}

type Result struct {
	// todo Rad versions
	Diagnostics []Diagnostic
}

type Pos struct {
	Line      int `json:"line"`      // Zero-indexed
	Character int `json:"character"` // Zero-indexed
}

type Range struct {
	Start Pos `json:"start"`
	End   Pos `json:"end"`
}

// todo mostly unused atm
type Opts struct {
	Errors bool
	Warns  bool
	Lints  bool
}

func NewOpts() Opts {
	return Opts{
		Errors: true,
		Warns:  true,
		Lints:  true,
	}
}
