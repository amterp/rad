package testing

import (
	"testing"

	"github.com/amterp/rad/core"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
)

func TestSeverityString(t *testing.T) {
	tests := []struct {
		severity core.Severity
		expected string
	}{
		{core.SeverityError, "error"},
		{core.SeverityWarning, "warning"},
		{core.SeverityNote, "note"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.severity.String())
		})
	}
}

func TestSpanLineAndColumn(t *testing.T) {
	span := rl.Span{
		File:     "test.rad",
		StartRow: 0,
		StartCol: 0,
		EndRow:   0,
		EndCol:   5,
	}

	// 0-indexed internally, 1-indexed for display
	assert.Equal(t, 1, span.StartLine())
	assert.Equal(t, 1, span.StartColumn())
}

func TestNewPrimaryLabel(t *testing.T) {
	span := rl.Span{File: "test.rad", StartRow: 5, StartCol: 10}
	label := core.NewPrimaryLabel(span, "error here")

	assert.True(t, label.Primary)
	assert.Equal(t, "error here", label.Message)
	assert.Equal(t, "test.rad", label.Span.File)
}

func TestNewSecondaryLabel(t *testing.T) {
	span := rl.Span{File: "test.rad", StartRow: 2, StartCol: 5}
	label := core.NewSecondaryLabel(span, "assigned here")

	assert.False(t, label.Primary)
	assert.Equal(t, "assigned here", label.Message)
}

func TestDiagnosticWithHint(t *testing.T) {
	span := rl.Span{File: "test.rad"}
	diag := core.NewDiagnostic(core.SeverityError, rl.ErrInvalidSyntax, "test error", "src", span)

	diag = diag.WithHint("try this instead")
	assert.Len(t, diag.Hints, 1)
	assert.Equal(t, "try this instead", diag.Hints[0])

	diag = diag.WithHints("hint 2", "hint 3")
	assert.Len(t, diag.Hints, 3)
}

func TestDiagnosticWithSecondaryLabel(t *testing.T) {
	span := rl.Span{File: "test.rad"}
	diag := core.NewDiagnostic(core.SeverityError, rl.ErrInvalidSyntax, "test error", "src", span)

	secondarySpan := rl.Span{File: "test.rad", StartRow: 1}
	diag = diag.WithSecondaryLabel(secondarySpan, "defined here")

	assert.Len(t, diag.Labels, 2)
	assert.True(t, diag.Labels[0].Primary)
	assert.False(t, diag.Labels[1].Primary)
	assert.Equal(t, "defined here", diag.Labels[1].Message)
}

func TestDiagnosticPrimarySpan(t *testing.T) {
	span := rl.Span{File: "primary.rad", StartRow: 10}
	diag := core.NewDiagnostic(core.SeverityError, rl.ErrInvalidSyntax, "test", "src", span)

	primary := diag.PrimarySpan()
	assert.NotNil(t, primary)
	assert.Equal(t, "primary.rad", primary.File)
	assert.Equal(t, 10, primary.StartRow)
}

func TestDiagnosticCollectorLimit(t *testing.T) {
	collector := core.NewDiagnosticCollectorWithLimit(3)

	span := rl.Span{File: "test.rad"}

	// Add diagnostics up to limit
	for i := 0; i < 5; i++ {
		diag := core.NewDiagnostic(core.SeverityError, rl.ErrInvalidSyntax, "error", "src", span)
		collector.Add(diag)
	}

	// Only 3 stored (the limit)
	assert.Equal(t, 3, collector.Count())
	// But 5 were emitted
	assert.Equal(t, 5, collector.TotalEmitted())
	// 2 remaining beyond limit
	assert.Equal(t, 2, collector.Remaining())
	assert.True(t, collector.AtLimit())
}

func TestDiagnosticCollectorAddReturnsFalseAtLimit(t *testing.T) {
	collector := core.NewDiagnosticCollectorWithLimit(2)
	span := rl.Span{File: "test.rad"}
	diag := core.NewDiagnostic(core.SeverityError, rl.ErrInvalidSyntax, "error", "src", span)

	assert.True(t, collector.Add(diag))  // 1st - ok
	assert.True(t, collector.Add(diag))  // 2nd - ok
	assert.False(t, collector.Add(diag)) // 3rd - at limit
	assert.False(t, collector.Add(diag)) // 4th - still at limit
}

func TestDiagnosticCollectorHasErrors(t *testing.T) {
	collector := core.NewDiagnosticCollector()
	span := rl.Span{File: "test.rad"}

	// Empty collector has no errors
	assert.False(t, collector.HasErrors())

	// Add a warning - still no errors
	warning := core.NewDiagnostic(core.SeverityWarning, rl.ErrInvalidSyntax, "warning", "src", span)
	collector.Add(warning)
	assert.False(t, collector.HasErrors())

	// Add an error - now has errors
	err := core.NewDiagnostic(core.SeverityError, rl.ErrInvalidSyntax, "error", "src", span)
	collector.Add(err)
	assert.True(t, collector.HasErrors())
}

func TestDiagnosticCollectorIsEmpty(t *testing.T) {
	collector := core.NewDiagnosticCollector()
	assert.True(t, collector.IsEmpty())

	span := rl.Span{File: "test.rad"}
	diag := core.NewDiagnostic(core.SeverityError, rl.ErrInvalidSyntax, "error", "src", span)
	collector.Add(diag)
	assert.False(t, collector.IsEmpty())
}

func TestNewDiagnosticFromCheck(t *testing.T) {
	// Create a check.Diagnostic
	code := rl.ErrMissingColon
	suggestion := "add a colon after the identifier"
	checkDiag := check.Diagnostic{
		OriginalSrc: "x = 5\ny = 10",
		Range: check.Range{
			Start: check.Pos{Line: 0, Character: 1},
			End:   check.Pos{Line: 0, Character: 2},
		},
		RangedSrc:  "=",
		LineSrc:    "x = 5",
		Severity:   check.Error,
		Message:    "missing colon",
		Code:       &code,
		Suggestion: &suggestion,
	}

	// Convert to core.Diagnostic
	coreDiag := core.NewDiagnosticFromCheck(checkDiag, "script.rad")

	assert.Equal(t, core.SeverityError, coreDiag.Severity)
	assert.Equal(t, rl.ErrMissingColon, coreDiag.Code)
	assert.Equal(t, "missing colon", coreDiag.Message)
	assert.Equal(t, "x = 5\ny = 10", coreDiag.Source)

	// Check span
	assert.Len(t, coreDiag.Labels, 1)
	assert.True(t, coreDiag.Labels[0].Primary)
	assert.Equal(t, "script.rad", coreDiag.Labels[0].Span.File)
	assert.Equal(t, 0, coreDiag.Labels[0].Span.StartRow)
	assert.Equal(t, 1, coreDiag.Labels[0].Span.StartCol)

	// Check suggestion became hint
	assert.Len(t, coreDiag.Hints, 1)
	assert.Equal(t, "add a colon after the identifier", coreDiag.Hints[0])
}

func TestNewDiagnosticFromCheckSeverityMapping(t *testing.T) {
	tests := []struct {
		checkSev check.Severity
		coreSev  core.Severity
	}{
		{check.Error, core.SeverityError},
		{check.Warning, core.SeverityWarning},
		{check.Hint, core.SeverityNote},
		{check.Info, core.SeverityNote},
	}

	for _, tt := range tests {
		t.Run(tt.checkSev.String(), func(t *testing.T) {
			checkDiag := check.Diagnostic{
				OriginalSrc: "test",
				Severity:    tt.checkSev,
				Message:     "test",
			}
			coreDiag := core.NewDiagnosticFromCheck(checkDiag, "test.rad")
			assert.Equal(t, tt.coreSev, coreDiag.Severity)
		})
	}
}

func TestNewDiagnosticFromCheckNilCode(t *testing.T) {
	// When check.Diagnostic has nil code, should use ErrGenericRuntime
	checkDiag := check.Diagnostic{
		OriginalSrc: "test",
		Severity:    check.Error,
		Message:     "test",
		Code:        nil,
	}
	coreDiag := core.NewDiagnosticFromCheck(checkDiag, "test.rad")
	assert.Equal(t, rl.ErrGenericRuntime, coreDiag.Code)
}
