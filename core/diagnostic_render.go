package core

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/amterp/color"
)

// DiagnosticRenderer renders diagnostics in Rust-style format.
type DiagnosticRenderer struct {
	writer   io.Writer
	useColor bool
}

// NewDiagnosticRenderer creates a renderer that writes to the given writer.
func NewDiagnosticRenderer(w io.Writer) *DiagnosticRenderer {
	return &DiagnosticRenderer{
		writer:   w,
		useColor: !color.NoColor,
	}
}

// Render renders a single diagnostic.
func (r *DiagnosticRenderer) Render(d Diagnostic) {
	r.renderHeader(d)
	r.renderLabels(d)
	r.renderHints(d)
	r.renderInfoLine(d)
	fmt.Fprintln(r.writer)
}

// RenderAll renders all diagnostics in a collector and displays truncation message if needed.
func (r *DiagnosticRenderer) RenderAll(c *DiagnosticCollector) {
	for _, d := range c.Diagnostics() {
		r.Render(d)
	}
	if remaining := c.Remaining(); remaining > 0 {
		fmt.Fprintf(r.writer, "%s\n", r.yellow(fmt.Sprintf("...and %d more errors", remaining)))
	}
}

// renderHeader renders: error[RAD10009]: message
func (r *DiagnosticRenderer) renderHeader(d Diagnostic) {
	var severityStr string
	switch d.Severity {
	case SeverityError:
		severityStr = r.red("error")
	case SeverityWarning:
		severityStr = r.yellow("warning")
	case SeverityNote:
		severityStr = r.cyan("note")
	}

	codeStr := r.red(fmt.Sprintf("[%s]", d.Code.String()))
	fmt.Fprintf(r.writer, "%s%s: %s\n", severityStr, codeStr, d.Message)
}

// renderLabels renders the source context with labels.
func (r *DiagnosticRenderer) renderLabels(d Diagnostic) {
	if len(d.Labels) == 0 || d.Source == "" {
		return
	}

	lines := strings.Split(d.Source, "\n")

	// Get the primary span for the location header
	primary := d.PrimarySpan()
	if primary == nil && len(d.Labels) > 0 {
		primary = &d.Labels[0].Span
	}

	// Render location: --> file:line:col
	if primary != nil {
		file := primary.File
		if file == "" {
			file = "<stdin>"
		}
		location := fmt.Sprintf("  --> %s:%d:%d", file, primary.StartLine(), primary.StartColumn())
		fmt.Fprintln(r.writer, r.blue(location))
	}

	// Sort labels by line number for rendering
	sortedLabels := make([]Label, len(d.Labels))
	copy(sortedLabels, d.Labels)
	sort.Slice(sortedLabels, func(i, j int) bool {
		return sortedLabels[i].Span.StartRow < sortedLabels[j].Span.StartRow
	})

	// Determine line number width for gutter
	maxLine := 0
	for _, label := range sortedLabels {
		if label.Span.EndRow > maxLine {
			maxLine = label.Span.EndRow
		}
	}
	gutterWidth := len(fmt.Sprintf("%d", maxLine+1)) // +1 for 1-indexing

	// Group labels by their start line
	labelsByLine := make(map[int][]Label)
	for _, label := range sortedLabels {
		labelsByLine[label.Span.StartRow] = append(labelsByLine[label.Span.StartRow], label)
	}

	// Determine which lines to show
	linesToShow := r.getLinesToShow(sortedLabels, len(lines))

	// Opening gutter
	fmt.Fprintf(r.writer, "%s\n", r.blue(strings.Repeat(" ", gutterWidth+1)+"|"))

	prevLine := -2 // Track for ellipsis
	for _, lineIdx := range linesToShow {
		// Show ellipsis if there's a gap
		if prevLine >= 0 && lineIdx > prevLine+1 {
			fmt.Fprintf(r.writer, "%s\n", r.blue("..."))
		}
		prevLine = lineIdx

		// Render the source line with gutter
		lineNum := lineIdx + 1 // 1-indexed for display
		r.renderSourceLine(lines, lineIdx, lineNum, gutterWidth)

		// Render any labels for this line
		if labels, ok := labelsByLine[lineIdx]; ok {
			r.renderLineLabels(labels, lines[lineIdx], gutterWidth)
		}
	}

	// Closing gutter
	fmt.Fprintf(r.writer, "%s\n", r.blue(strings.Repeat(" ", gutterWidth+1)+"|"))
}

// getLinesToShow determines which lines to display based on labels.
// Shows 1 line before and 2 lines after each labeled line, with context merging.
func (r *DiagnosticRenderer) getLinesToShow(labels []Label, totalLines int) []int {
	lineSet := make(map[int]bool)

	for _, label := range labels {
		// Add the labeled line and context
		start := label.Span.StartRow - 1 // 1 line before
		end := label.Span.StartRow + 2   // 2 lines after

		if start < 0 {
			start = 0
		}
		if end >= totalLines {
			end = totalLines - 1
		}

		for i := start; i <= end; i++ {
			lineSet[i] = true
		}
	}

	// Convert to sorted slice
	var lines []int
	for line := range lineSet {
		lines = append(lines, line)
	}
	sort.Ints(lines)

	return lines
}

// renderSourceLine renders a single source line with line number gutter.
func (r *DiagnosticRenderer) renderSourceLine(lines []string, lineIdx, lineNum, gutterWidth int) {
	if lineIdx < 0 || lineIdx >= len(lines) {
		return
	}

	line := lines[lineIdx]
	// Truncate long lines
	if len(line) > 120 {
		line = line[:117] + "..."
	}

	gutter := fmt.Sprintf("%*d", gutterWidth, lineNum)
	fmt.Fprintf(r.writer, "%s %s\n", r.blue(gutter+" |"), line)
}

// renderLineLabels renders the underline and message for labels on a line.
func (r *DiagnosticRenderer) renderLineLabels(labels []Label, line string, gutterWidth int) {
	// Build the underline string
	underline := make([]rune, len(line)+10) // Extra space for potential overflow
	for i := range underline {
		underline[i] = ' '
	}

	// Track which label's message to show (rightmost for now)
	var messageLabel *Label
	messageCol := -1

	for i := range labels {
		label := &labels[i]
		startCol := label.Span.StartCol
		endCol := label.Span.EndCol

		// Handle same-line spans
		if label.Span.StartRow == label.Span.EndRow {
			// Ensure we don't go past the line length
			if endCol > len(line) {
				endCol = len(line)
			}
			if startCol > len(line) {
				startCol = len(line)
			}
		} else {
			// Multi-line span: underline to end of this line
			endCol = len(line)
		}

		// Fill in the underline characters
		char := '^'
		if !label.Primary {
			char = '-'
		}
		for col := startCol; col < endCol && col < len(underline); col++ {
			underline[col] = char
		}

		// Track the label with a message to display
		if label.Message != "" && startCol > messageCol {
			messageLabel = label
			messageCol = startCol
		}
	}

	// Trim trailing spaces
	underlineStr := strings.TrimRight(string(underline), " ")
	if underlineStr == "" {
		return
	}

	// Colorize the underline
	var coloredUnderline string
	if labels[0].Primary {
		coloredUnderline = r.red(underlineStr)
	} else {
		coloredUnderline = r.blue(underlineStr)
	}

	// Build the output line
	gutter := r.blue(strings.Repeat(" ", gutterWidth) + " |")

	if messageLabel != nil && messageLabel.Message != "" {
		// Add message after the underline
		fmt.Fprintf(r.writer, "%s %s %s\n", gutter, coloredUnderline, r.colorForLabel(messageLabel, messageLabel.Message))
	} else {
		fmt.Fprintf(r.writer, "%s %s\n", gutter, coloredUnderline)
	}
}

// renderHints renders the help hints.
func (r *DiagnosticRenderer) renderHints(d Diagnostic) {
	for _, hint := range d.Hints {
		fmt.Fprintf(r.writer, "   %s %s\n", r.green("= help:"), hint)
	}
}

// renderInfoLine renders the "rad explain" info line.
func (r *DiagnosticRenderer) renderInfoLine(d Diagnostic) {
	if d.Code != "" {
		info := fmt.Sprintf("= info: rad explain %s", d.Code.String())
		fmt.Fprintf(r.writer, "   %s\n", r.cyan(info))
	}
}

// Color helper methods
func (r *DiagnosticRenderer) red(s string) string {
	if !r.useColor {
		return s
	}
	return color.RedString(s)
}

func (r *DiagnosticRenderer) yellow(s string) string {
	if !r.useColor {
		return s
	}
	return color.YellowString(s)
}

func (r *DiagnosticRenderer) blue(s string) string {
	if !r.useColor {
		return s
	}
	return color.BlueString(s)
}

func (r *DiagnosticRenderer) green(s string) string {
	if !r.useColor {
		return s
	}
	return color.GreenString(s)
}

func (r *DiagnosticRenderer) cyan(s string) string {
	if !r.useColor {
		return s
	}
	return color.CyanString(s)
}

func (r *DiagnosticRenderer) colorForLabel(label *Label, s string) string {
	if label.Primary {
		return r.red(s)
	}
	return r.blue(s)
}
