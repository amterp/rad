package testing

import (
	"bytes"
	"os"
	"testing"

	"github.com/amterp/color"
	"github.com/amterp/rad/core"
	"github.com/amterp/rad/rts/rl"
	"github.com/stretchr/testify/assert"
)

func init() {
	// Disable colors for deterministic test output
	color.NoColor = true
}

func TestRenderSingleSpanError(t *testing.T) {
	// Ensure colors are disabled for this test
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	source := `args:
    name str = $username`

	span := rl.Span{
		File:     "script.rad",
		StartRow: 1,
		StartCol: 15,
		EndRow:   1,
		EndCol:   24,
	}

	diag := core.NewDiagnostic(
		core.SeverityError,
		rl.ErrUnexpectedToken,
		"unexpected token",
		source,
		span,
	).WithHint("variable names cannot start with '$'")

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.Render(diag)

	output := buf.String()

	// Check key parts of the output
	assert.Contains(t, output, "error[RAD10009]")
	assert.Contains(t, output, "unexpected token")
	assert.Contains(t, output, "--> script.rad:2:16")
	assert.Contains(t, output, "name str = $username")
	assert.Contains(t, output, "^^^^^^^^^")
	assert.Contains(t, output, "= help:")
	assert.Contains(t, output, "variable names cannot start with '$'")
	assert.Contains(t, output, "= info: rad explain RAD10009")
}

func TestRenderWarning(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	source := `x = 5`
	span := rl.Span{
		File:     "test.rad",
		StartRow: 0,
		StartCol: 0,
		EndRow:   0,
		EndCol:   1,
	}

	diag := core.NewDiagnostic(
		core.SeverityWarning,
		rl.ErrGenericRuntime,
		"unused variable",
		source,
		span,
	)

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.Render(diag)

	output := buf.String()
	assert.Contains(t, output, "warning[RAD20000]")
	assert.Contains(t, output, "unused variable")
}

func TestRenderMultiSpan(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	source := `count = "five"
result = count * factor`

	// Primary span: where the error occurs
	primarySpan := rl.Span{
		File:     "script.rad",
		StartRow: 1,
		StartCol: 9,
		EndRow:   1,
		EndCol:   14,
	}

	// Secondary span: where the value was defined
	secondarySpan := rl.Span{
		File:     "script.rad",
		StartRow: 0,
		StartCol: 8,
		EndRow:   0,
		EndCol:   14,
	}

	diag := core.NewDiagnosticWithLabels(
		core.SeverityError,
		rl.ErrGenericRuntime,
		"cannot multiply string by int",
		source,
		[]core.Label{
			core.NewPrimaryLabel(primarySpan, "expected int, found string"),
			core.NewSecondaryLabel(secondarySpan, "assigned as string here"),
		},
	)

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.Render(diag)

	output := buf.String()

	// Should show both spans with labels
	assert.Contains(t, output, "cannot multiply string by int")
	assert.Contains(t, output, `count = "five"`)
	assert.Contains(t, output, `result = count * factor`)
}

func TestRenderLongLinesTruncated(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	// Create a very long line
	longLine := "x = " + string(make([]byte, 200)) // 200+ chars
	for i := 4; i < len(longLine); i++ {
		longLine = longLine[:i] + "a" + longLine[i+1:]
	}

	span := rl.Span{
		File:     "test.rad",
		StartRow: 0,
		StartCol: 0,
		EndRow:   0,
		EndCol:   5,
	}

	diag := core.NewDiagnostic(
		core.SeverityError,
		rl.ErrInvalidSyntax,
		"test error",
		longLine,
		span,
	)

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.Render(diag)

	output := buf.String()

	// Line should be truncated with ...
	assert.Contains(t, output, "...")
	// Should not contain the full 200 char line
	assert.Less(t, len(output), 500)
}

func TestRenderCollectorWithRemaining(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	collector := core.NewDiagnosticCollectorWithLimit(2)
	span := rl.Span{File: "test.rad"}

	// Add 5 diagnostics (only 2 will be stored)
	for i := 0; i < 5; i++ {
		diag := core.NewDiagnostic(
			core.SeverityError,
			rl.ErrInvalidSyntax,
			"error",
			"x = 1",
			span,
		)
		collector.Add(diag)
	}

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.RenderAll(collector)

	output := buf.String()

	// Should show truncation message
	assert.Contains(t, output, "...and 3 more errors")
}

func TestRenderNote(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	source := `x = 5`
	span := rl.Span{
		File:     "test.rad",
		StartRow: 0,
		StartCol: 0,
		EndRow:   0,
		EndCol:   1,
	}

	diag := core.NewDiagnostic(
		core.SeverityNote,
		rl.ErrGenericRuntime,
		"related definition here",
		source,
		span,
	)

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.Render(diag)

	output := buf.String()
	assert.Contains(t, output, "note[RAD20000]")
}

func TestRenderEmptySource(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	span := rl.Span{File: "test.rad"}
	diag := core.NewDiagnostic(
		core.SeverityError,
		rl.ErrInvalidSyntax,
		"error with no source",
		"", // empty source
		span,
	)

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)

	// Should not panic
	renderer.Render(diag)

	output := buf.String()
	assert.Contains(t, output, "error[RAD10001]")
	assert.Contains(t, output, "error with no source")
}

func TestRenderStdinFile(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	source := `x = 5`
	span := rl.Span{
		File:     "", // empty file means stdin
		StartRow: 0,
		StartCol: 0,
		EndRow:   0,
		EndCol:   1,
	}

	diag := core.NewDiagnostic(
		core.SeverityError,
		rl.ErrInvalidSyntax,
		"error",
		source,
		span,
	)

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.Render(diag)

	output := buf.String()
	assert.Contains(t, output, "<stdin>:1:1")
}

// Verify the overall format matches the spec
func TestRenderFormatMatchesSpec(t *testing.T) {
	oldNoColor := color.NoColor
	color.NoColor = true
	defer func() { color.NoColor = oldNoColor }()

	source := `args:
    name str = $username`

	span := rl.Span{
		File:     "script.rad",
		StartRow: 1,
		StartCol: 15,
		EndRow:   1,
		EndCol:   24,
	}

	diag := core.NewDiagnostic(
		core.SeverityError,
		rl.ErrUnexpectedToken,
		"unexpected token",
		source,
		span,
	).WithHint("variable names cannot start with '$'")

	var buf bytes.Buffer
	renderer := core.NewDiagnosticRenderer(&buf)
	renderer.Render(diag)

	output := buf.String()

	// Verify structural elements from the spec:
	// 1. Header line with severity, code, and message
	assert.Contains(t, output, "error[RAD10009]: unexpected token")

	// 2. Location line
	assert.Contains(t, output, "--> script.rad:2:16")

	// 3. Gutter with pipe
	assert.Contains(t, output, "|")

	// 4. Source line with line number
	assert.Contains(t, output, "2 |")
	assert.Contains(t, output, "name str = $username")

	// 5. Help line
	assert.Contains(t, output, "= help:")

	// 6. Info line
	assert.Contains(t, output, "= info: rad explain")
}

func TestMain(m *testing.M) {
	// Ensure colors are off for all tests
	color.NoColor = true
	os.Exit(m.Run())
}
