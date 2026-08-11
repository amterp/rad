package core

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/amterp/color"
	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/rl"
)

// Layout constants for diagnostic rendering.
const (
	// diagWidthFloor is the narrowest terminal we lay out for. Below it, text
	// is going to look cramped whatever we do, and shrinking further just
	// produces one-word lines.
	diagWidthFloor = 40
	// diagWidthCap bounds prose regardless of how wide the terminal is. Long
	// measures are hard to read, and a 1900-column terminal should not produce
	// a 1900-column sentence.
	diagWidthCap = 100

	// cutMarker flags a horizontally trimmed source line. Deliberately ASCII
	// rather than the "…" used by table truncation: it means "the line
	// continues", not "content was dropped", and a fixed 3-column width keeps
	// the caret offset arithmetic exact on terminals of any encoding.
	cutMarker  = "..."
	cutMarkerW = 3

	// spanContext is how many columns of lead-in to keep before the span when
	// a line has to be trimmed from the left.
	spanContext = 6
	// indentElideAt / indentKeep: indentation deeper than this is collapsed to
	// this, since nesting depth reads fine from a few columns and the rest is
	// just pushing the code out of view.
	indentElideAt = 20
	indentKeep    = 4
	// minSourceCols is the least source room worth windowing into. Narrower
	// than this, trimming costs more context than it buys.
	minSourceCols = 20
	// minMessageCols is the least room a wrapped label message needs before we
	// give up its hanging indent and start it at the gutter.
	minMessageCols = 24
)

// DiagnosticRenderer renders diagnostics in Rust-style format, laid out to fit
// the terminal.
type DiagnosticRenderer struct {
	writer   io.Writer
	useColor bool

	// prose bounds the header, the "= ..." lines and label messages.
	prose int
	// source bounds the "|" block. Unlike prose it is not capped: a wider
	// terminal should never show less code.
	source int
}

// NewDiagnosticRenderer creates a renderer that writes to the given writer,
// laid out for the current terminal.
func NewDiagnosticRenderer(w io.Writer) *DiagnosticRenderer {
	return NewDiagnosticRendererWidth(w, GetTermWidth())
}

// NewDiagnosticRendererWidth creates a renderer for an explicit terminal width.
func NewDiagnosticRendererWidth(w io.Writer, termWidth int) *DiagnosticRenderer {
	termWidth = usableTermWidth(termWidth)
	return &DiagnosticRenderer{
		writer:   w,
		useColor: !color.NoColor,
		prose:    com.IntMin(termWidth, diagWidthCap),
		source:   termWidth,
	}
}

// usableTermWidth turns a raw terminal width into one worth laying out for.
//
// GetTermWidth returns a huge sentinel when there is no terminal to measure.
// That is right for tables - piping data into a file shouldn't truncate it -
// but wrong for error output: a diagnostic redirected to a file or a pager is
// still meant to be read, so lay it out as a default terminal would. It also
// makes piped output deterministic, which the tests rely on.
func usableTermWidth(termWidth int) int {
	if termWidth >= defaultTermWidth {
		return diagWidthCap
	}
	return com.IntMax(termWidth, diagWidthFloor)
}

// DiagnosticProseWidth reports the column budget for prose in error output.
// Anything hand-formatting an error message should wrap to this rather than
// picking its own width, so every error rad prints fits the same terminal.
func DiagnosticProseWidth() int {
	return com.IntMin(usableTermWidth(GetTermWidth()), diagWidthCap)
}

// RenderDiagnosticToString renders d to a string, honoring the current color
// setting and terminal width. No trailing newline.
func RenderDiagnosticToString(d Diagnostic) string {
	var sb strings.Builder
	NewDiagnosticRenderer(&sb).renderBody(d)
	return strings.TrimRight(sb.String(), "\n")
}

// Render renders a single diagnostic.
func (r *DiagnosticRenderer) Render(d Diagnostic) {
	r.renderBody(d)
	fmt.Fprintln(r.writer)
}

func (r *DiagnosticRenderer) renderBody(d Diagnostic) {
	// The gutter width sets the column everything else hangs from, including
	// the "= ..." lines below the snippet, so it is resolved once up front.
	snippet := r.snippetFor(d)

	r.renderHeader(d)
	r.renderLabels(d, snippet)
	r.renderCallStack(d, snippet.gutterWidth)
	r.renderHints(d, snippet.gutterWidth)
	r.renderInfoLine(d, snippet.gutterWidth)
}

// snippet is the resolved source context for a diagnostic: which lines to
// show, how wide their numbers are, and the horizontal window they share.
type snippet struct {
	lines       []string
	show        []int
	gutterWidth int
	window      srcWindow
}

func (r *DiagnosticRenderer) snippetFor(d Diagnostic) snippet {
	if len(d.Labels) == 0 || d.Source == "" {
		return snippet{gutterWidth: 1}
	}

	lines := strings.Split(d.Source, "\n")
	for i := range lines {
		lines[i] = strings.TrimSuffix(lines[i], "\r")
	}

	sorted := make([]Label, len(d.Labels))
	copy(sorted, d.Labels)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Span.StartRow < sorted[j].Span.StartRow
	})

	show := r.getLinesToShow(sorted, len(lines))

	// Context lines can have higher numbers than any labeled line, so measure
	// against what is actually displayed.
	maxDisplayLine := 0
	if len(show) > 0 {
		maxDisplayLine = show[len(show)-1]
	}
	gutterWidth := len(fmt.Sprintf("%d", maxDisplayLine+1)) // +1 for 1-indexing

	primary := d.PrimarySpan()
	if primary == nil {
		primary = &d.Labels[0].Span
	}

	return snippet{
		lines:       lines,
		show:        show,
		gutterWidth: gutterWidth,
		window:      r.chooseWindow(lines, *primary, gutterWidth),
	}
}

// RenderAll renders all diagnostics in a collector and displays truncation message if needed.
func (r *DiagnosticRenderer) RenderAll(c *DiagnosticCollector) {
	for _, d := range c.Diagnostics() {
		r.Render(d)
	}
	if remaining := c.Remaining(); remaining > 0 {
		r.line("%s", r.yellow(fmt.Sprintf("...and %d more errors", remaining)))
	}
}

// line writes one output line, right-trimmed. Trailing whitespace would
// otherwise force go-snap to store diagnostic snapshots as quoted strings,
// making every one of them unreadable.
func (r *DiagnosticRenderer) line(format string, a ...any) {
	fmt.Fprintln(r.writer, strings.TrimRight(fmt.Sprintf(format, a...), " \t"))
}

// renderHeader renders: error[RAD10009]: message
func (r *DiagnosticRenderer) renderHeader(d Diagnostic) {
	prefix := d.Severity.String()
	if d.Code != "" {
		prefix += fmt.Sprintf("[%s]", d.Code.String())
	}

	// Continuation lines get a small fixed indent rather than aligning under
	// the message: a hanging indent past "error[RAD40023]: " would eat 17 of
	// the 40 columns we guarantee.
	lines := com.WrapPrefixed(d.Message, prefix+": ", "  ", r.prose)

	for i, l := range lines {
		if i == 0 {
			// Only the prefix is colored; the message keeps the default
			// foreground so it stays readable on any theme.
			r.line("%s%s", r.severityColor(d.Severity, prefix+":"), strings.TrimPrefix(l, prefix+":"))
		} else {
			r.line("%s", l)
		}
	}
}

// renderLabels renders the source context with labels.
func (r *DiagnosticRenderer) renderLabels(d Diagnostic, s snippet) {
	if len(s.show) == 0 {
		return
	}

	primary := d.PrimarySpan()
	if primary == nil {
		primary = &d.Labels[0].Span
	}
	r.renderLocation(*primary)

	labelsByLine := make(map[int][]Label)
	for _, label := range d.Labels {
		labelsByLine[label.Span.StartRow] = append(labelsByLine[label.Span.StartRow], label)
	}

	bar := r.blue(strings.Repeat(" ", s.gutterWidth+1) + "|")
	r.line("%s", bar)

	prevLine := -2
	for _, lineIdx := range s.show {
		if prevLine >= 0 && lineIdx > prevLine+1 {
			// Gutter-aligned so it can't be confused with cutMarker, which
			// means something different and appears on the same lines.
			r.line("%s", r.blue(fmt.Sprintf("%-*s|", s.gutterWidth+1, "...")))
		}
		prevLine = lineIdx

		r.renderSourceLine(s.lines[lineIdx], lineIdx+1, s.gutterWidth, s.window)

		if labels, ok := labelsByLine[lineIdx]; ok {
			r.renderLineLabels(labels, s.lines[lineIdx], s.gutterWidth, s.window)
		}
	}

	r.line("%s", bar)
}

// renderLocation renders "  --> file:line:col", shortening the path from the
// left rather than wrapping - a split location is neither readable nor
// clickable.
func (r *DiagnosticRenderer) renderLocation(span rl.Span) {
	file := span.File
	if file == "" {
		file = "<stdin>"
	}

	suffix := fmt.Sprintf(":%d:%d", span.StartLine(), span.StartColumn())
	const lead = "  --> "

	if budget := r.prose - len(lead) - com.DisplayWidth(suffix); com.DisplayWidth(file) > budget {
		file = com.ShortenPathLeft(file, budget)
	}

	r.line("%s", r.blue(lead+file+suffix))
}

// srcWindow is the horizontal slice of the source lines being displayed.
// Columns are display columns of the tab-expanded line.
type srcWindow struct {
	left     int  // first column shown
	capacity int  // columns available for text, after any left marker
	cutL     bool // text was trimmed on the left
}

// forLine resolves the window against one specific line, which may be shorter
// than the line the window was chosen for.
func (w srcWindow) forLine(lineWidth int) (left, right int, cutL, cutR bool) {
	if lineWidth <= w.left {
		// This line ends before the window starts - nothing of it is in view.
		return w.left, w.left, false, false
	}
	right = lineWidth
	if right-w.left > w.capacity {
		cutR = true
		right = w.left + w.capacity - cutMarkerW
	}
	return w.left, right, w.cutL, cutR
}

// chooseWindow picks the horizontal window for a snippet, from the line holding
// the primary span. The span's first column must end up visible; beyond that we
// prefer showing the line from its start, and prefer collapsing deep
// indentation over cutting into code.
func (r *DiagnosticRenderer) chooseWindow(lines []string, primary rl.Span, gutterWidth int) srcWindow {
	avail := r.source - (gutterWidth + 3)

	full := srcWindow{left: 0, capacity: com.IntMax(avail, 1), cutL: false}
	if avail < minSourceCols || primary.StartRow < 0 || primary.StartRow >= len(lines) {
		return full
	}

	raw := lines[primary.StartRow]
	expanded, colOf := com.ExpandTabs(raw)
	lineWidth := com.DisplayWidth(expanded)
	if lineWidth <= avail {
		return full
	}

	spanL := columnAt(colOf, primary.StartCol)
	spanR := columnAt(colOf, primary.EndCol)
	if primary.EndRow != primary.StartRow {
		spanR = lineWidth
	}
	indentWidth := lineWidth - com.DisplayWidth(strings.TrimLeft(expanded, " "))

	window := func(left int) srcWindow {
		capacity := avail
		if left > 0 {
			capacity -= cutMarkerW
		}
		return srcWindow{left: left, capacity: capacity, cutL: left > 0}
	}

	// Candidates in order of preference: collapse deep indentation, else show
	// the line from its start, else anchor just before the span.
	var candidates []int
	if indentWidth >= indentElideAt && spanL >= indentWidth {
		candidates = append(candidates, indentWidth-indentKeep)
	}
	candidates = append(candidates, 0, com.IntMax(0, spanL-spanContext))

	// A windowed line nearly always has a right cut, so budget for one.
	lastVisible := func(left int) int {
		return left + window(left).capacity - cutMarkerW
	}

	// Two passes. Showing the whole span is what makes the underline mean what
	// it says, so it wins over showing the line from column zero; only when no
	// window can hold the span - it is wider than the terminal - do we settle
	// for having its start in view and let the carets clip.
	for _, wholeSpan := range []bool{true, false} {
		for _, left := range candidates {
			if left > spanL {
				continue
			}
			if wholeSpan && spanR <= lastVisible(left) {
				return window(left)
			}
			if !wholeSpan && spanL < lastVisible(left) {
				return window(left)
			}
		}
	}

	return window(com.IntMax(0, spanL-spanContext))
}

// renderSourceLine renders a single source line with line number gutter.
func (r *DiagnosticRenderer) renderSourceLine(raw string, lineNum, gutterWidth int, win srcWindow) {
	expanded, _ := com.ExpandTabs(raw)
	left, right, cutL, cutR := win.forLine(com.DisplayWidth(expanded))

	text := com.SliceColumns(expanded, left, right)
	if cutL {
		text = cutMarker + text
	}
	if cutR {
		text += cutMarker
	}

	gutter := fmt.Sprintf("%*d", gutterWidth, lineNum)
	r.line("%s %s", r.blue(gutter+" |"), text)
}

// renderLineLabels renders the underline and message for labels on a line.
func (r *DiagnosticRenderer) renderLineLabels(labels []Label, raw string, gutterWidth int, win srcWindow) {
	expanded, colOf := com.ExpandTabs(raw)
	lineWidth := com.DisplayWidth(expanded)
	left, right, cutL, _ := win.forLine(lineWidth)

	// One column past the end of the line, so an insertion point at end-of-line
	// - or on an empty line - still gets a caret. Bounded by the window's
	// capacity, which the underline may fill completely since it carries no
	// right-hand cut marker of its own.
	markWidth := com.IntMin(right-left+1, win.capacity)
	if markWidth <= 0 {
		return
	}

	marks := make([]rune, markWidth)
	owner := make([]*Label, markWidth)
	for i := range marks {
		marks[i] = ' '
	}

	// The label whose message we show, and the column it hangs from.
	var messageLabel *Label
	messageCol := -1

	for i := range labels {
		label := &labels[i]

		startCol := columnAt(colOf, label.Span.StartCol)
		endCol := columnAt(colOf, label.Span.EndCol)
		if label.Span.StartRow != label.Span.EndRow {
			// Multi-line span: underline to the end of this line.
			endCol = lineWidth
		}
		if endCol <= startCol {
			// Zero-width span (an insertion point) still gets one caret.
			endCol = startCol + 1
		}

		char := '^'
		if !label.Primary {
			char = '-'
		}
		for col := com.IntMax(startCol, left); col < com.IntMin(endCol, left+markWidth); col++ {
			marks[col-left] = char
			owner[col-left] = label
		}

		if label.Message != "" && startCol > messageCol {
			messageLabel = label
			messageCol = startCol
		}
	}

	underline := strings.TrimRight(string(marks), " ")
	if underline == "" {
		return
	}

	// The left marker occupies real columns on the source line, so the caret
	// row has to be pushed by the same amount to stay under its span.
	markerPad := 0
	if cutL {
		markerPad = cutMarkerW
	}

	var colored strings.Builder
	var currentLabel *Label
	var segment strings.Builder
	for i, ch := range []rune(underline) {
		if owner[i] != currentLabel && segment.Len() > 0 {
			colored.WriteString(r.colorSegment(segment.String(), currentLabel))
			segment.Reset()
		}
		currentLabel = owner[i]
		segment.WriteRune(ch)
	}
	if segment.Len() > 0 {
		colored.WriteString(r.colorSegment(segment.String(), currentLabel))
	}

	gutter := r.blue(strings.Repeat(" ", gutterWidth) + " |")
	pad := strings.Repeat(" ", markerPad)

	if messageLabel == nil || messageLabel.Message == "" {
		r.line("%s %s%s", gutter, pad, colored.String())
		return
	}

	// Message text is prose, so it answers to the prose cap even though it sits
	// inside the source block. Everything here budgets against the gutter the
	// line is actually printed with.
	textWidth := r.prose - (gutterWidth + 3)
	tail := markerPad + com.DisplayWidth(underline)

	if tail+1+com.DisplayWidth(messageLabel.Message) <= textWidth {
		r.line("%s %s%s %s", gutter, pad, colored.String(),
			r.colorForLabel(messageLabel, messageLabel.Message))
		return
	}

	// Too long to sit beside the underline, so it goes below it, hanging from
	// the span it describes.
	r.line("%s %s%s", gutter, pad, colored.String())

	indent := markerPad + com.IntMax(messageCol-left, 0)
	if textWidth-indent < minMessageCols {
		indent = 0
	}
	for _, l := range com.Wrap(messageLabel.Message, textWidth-indent) {
		r.line("%s %s%s", gutter, strings.Repeat(" ", indent), r.colorForLabel(messageLabel, l))
	}
}

// columnAt maps a byte offset within a line to its display column, clamping to
// the line's bounds.
func columnAt(colOf []int, byteOffset int) int {
	if byteOffset < 0 {
		return 0
	}
	if byteOffset >= len(colOf) {
		return colOf[len(colOf)-1]
	}
	return colOf[byteOffset]
}

// getLinesToShow determines which lines to display based on labels.
// Shows 1 line before and 2 lines after each labeled line, with context merging.
func (r *DiagnosticRenderer) getLinesToShow(labels []Label, totalLines int) []int {
	lineSet := make(map[int]bool)

	for _, label := range labels {
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

	var lines []int
	for line := range lineSet {
		lines = append(lines, line)
	}
	sort.Ints(lines)

	return lines
}

// colorSegment colors a segment of underline based on its owning label.
func (r *DiagnosticRenderer) colorSegment(s string, label *Label) string {
	if label == nil {
		return s // Spaces between labels, no color
	}
	if label.Primary {
		return r.red(s)
	}
	return r.blue(s)
}

// renderHints renders the help hints.
func (r *DiagnosticRenderer) renderHints(d Diagnostic, gutterWidth int) {
	for _, hint := range d.Hints {
		r.renderTagged("= help:", hint, r.green, gutterWidth)
	}
}

// renderInfoLine renders the "rad docs" info line pointing at the
// error code's documentation.
func (r *DiagnosticRenderer) renderInfoLine(d Diagnostic, gutterWidth int) {
	if d.Code != "" {
		r.renderTagged("= info:", "rad docs "+d.Code.String(), r.cyan, gutterWidth)
	}
}

// renderTagged renders a "  = tag: text" line, wrapping text with a hanging
// indent aligned under it. The "=" sits in the same column as the snippet's
// "|", so the whole block reads as one left edge.
func (r *DiagnosticRenderer) renderTagged(
	tag, text string,
	colorize func(string) string,
	gutterWidth int,
) {
	lead := strings.Repeat(" ", gutterWidth+1)
	first := lead + tag + " "
	cont := strings.Repeat(" ", com.DisplayWidth(first))

	for i, l := range com.WrapPrefixed(text, first, cont, r.prose) {
		if i == 0 {
			r.line("%s%s%s", lead, colorize(tag), strings.TrimPrefix(l, lead+tag))
		} else {
			r.line("%s", l)
		}
	}
}

// renderCallStack renders the call stack if present.
func (r *DiagnosticRenderer) renderCallStack(d Diagnostic, gutterWidth int) {
	if len(d.CallStack) == 0 {
		return
	}

	lead := strings.Repeat(" ", gutterWidth+1)
	r.line("%s%s", lead, r.cyan("= stack:"))
	for i, frame := range d.CallStack {
		indent := lead + "  "
		prefix := "in"
		if i == 0 {
			prefix = "at"
		}

		location := ""
		if frame.CallSite != nil {
			location = fmt.Sprintf("%s:%d:%d",
				frame.CallSite.File,
				frame.CallSite.StartLine(),
				frame.CallSite.StartColumn())
		}

		if location == "" {
			r.line("%s%s %s", indent, prefix, r.blue(frame.FunctionName))
			continue
		}

		// A frame is one unit - wrapping it would be unreadable. Shorten the
		// path instead, since the tail is the informative end.
		budget := r.prose - com.DisplayWidth(indent+prefix+" "+frame.FunctionName+" ()")
		location = com.ShortenPathLeft(location, budget)
		r.line("%s%s %s (%s)", indent, prefix, r.blue(frame.FunctionName), location)
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

func (r *DiagnosticRenderer) white(s string) string {
	if !r.useColor {
		return s
	}
	return color.WhiteString(s)
}

// severityColor colors the header's "severity[CODE]:" prefix. The code shares
// the severity's color rather than being unconditionally red, which used to
// make warnings look like errors at a glance.
func (r *DiagnosticRenderer) severityColor(s Severity, text string) string {
	switch s {
	case SeverityError:
		return r.red(text)
	case SeverityWarning:
		return r.yellow(text)
	case SeverityNote:
		return r.cyan(text)
	case SeverityHint:
		return r.white(text)
	}
	return text
}

func (r *DiagnosticRenderer) colorForLabel(label *Label, s string) string {
	if label.Primary {
		return r.red(s)
	}
	return r.blue(s)
}
