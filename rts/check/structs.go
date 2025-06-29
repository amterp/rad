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
		return "Warning"
	case Error:
		return "Error"
	default:
		return "Unknown"
	}
}

type Diagnostic struct {
	OriginalSrc string // Complete original src
	Range       Range
	RangedSrc   string // Src for just the Range
	LineSrc     string // Src for the line at the start of Range
	Severity    Severity
	Message     string
	Code        *rl.Error
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
