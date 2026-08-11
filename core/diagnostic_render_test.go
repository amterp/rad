package core

import (
	"strings"
	"testing"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/rl"
)

// renderAt renders d at an explicit terminal width, with color off.
func renderAt(d Diagnostic, termWidth int) []string {
	var sb strings.Builder
	r := NewDiagnosticRendererWidth(&sb, termWidth)
	r.useColor = false
	r.renderBody(d)
	return strings.Split(strings.TrimRight(sb.String(), "\n"), "\n")
}

func diag(source, message string, startRow, startCol, endCol int) Diagnostic {
	return NewDiagnostic(SeverityError, rl.ErrGenericRuntime, message, source, rl.Span{
		File:     "script.rad",
		StartRow: startRow,
		StartCol: startCol,
		EndRow:   startRow,
		EndCol:   endCol,
	})
}

const longMessage = "This value is inside quotes, but Rad already quotes interpolated " +
	"values in shell commands - these quotes will end up in the argument itself."

// renderCorpus is the awkward corner of the input space: things that have
// historically broken caret alignment or blown past the terminal edge.
func renderCorpus() map[string]Diagnostic {
	long := "code, out, err = quiet $`rad src/sources/reddit.rad --subreddits \"{subreddits}\" " +
		"--queries \"{queries}\" --lookback-hours {lookback} --limit {limit}` catch:"

	return map[string]Diagnostic{
		"short":                    diag("x = 1", "bad thing", 0, 4, 5),
		"long message":             diag("x = 1", longMessage, 0, 4, 5),
		"long line":                diag(long, "quoted interpolation", 0, 60, 72),
		"span at line end":         diag(long, "trailing", 0, len(long)-6, len(long)),
		"span wider than terminal": diag(long, "the whole command", 0, 25, len(long)),
		"zero width span":          diag("x = 1", "insert here", 0, 5, 5),
		"deep indent": diag(strings.Repeat(" ", 40)+"value = compute(a, b, c)",
			"deeply nested", 0, 48, 55),
		"tabs":       diag("\t\tyes no", "undefined identifier 'no'", 0, 6, 8),
		"cjk source": diag(`msg = "日本語のテキスト" + broken`, "type mismatch", 0, 28, 34),
		"cjk message": diag("x = 1",
			"日本語のメッセージです。これは長いメッセージで折り返しが必要になります。", 0, 4, 5),
		"empty source line": diag("x = 1\n\ny = 2", "blank", 1, 0, 0),
		"no source":         NewDiagnostic(SeverityError, rl.ErrGenericRuntime, longMessage, "", rl.Span{}),
		"multi line span": NewDiagnostic(SeverityError, rl.ErrGenericRuntime, "unterminated", "a = [1,\n2,\n3", rl.Span{
			File: "script.rad", StartRow: 0, StartCol: 4, EndRow: 2, EndCol: 1,
		}),
		"with hint": diag("x = 1", "bad thing", 0, 4, 5).
			WithHint("Reorder the targets to match, or name them all so assignment goes " +
				"by name - e.g. `stdout, code = $`cmd``."),
		"labelled": NewDiagnosticWithLabels(SeverityError, rl.ErrGenericRuntime, "mismatch", long,
			[]Label{
				NewPrimaryLabel(rl.Span{File: "s.rad", StartRow: 0, StartCol: 60, EndRow: 0, EndCol: 72},
					longMessage),
			}),
		"long path": NewDiagnostic(SeverityError, rl.ErrGenericRuntime, "bad", "x = 1", rl.Span{
			File:     "/Users/someone/src/some-project/deeply/nested/directory/script.rad",
			StartRow: 0, StartCol: 4, EndRow: 0, EndCol: 5,
		}),
	}
}

// The invariant the whole change exists to establish: nothing rad prints may
// exceed the terminal, at any width, for any diagnostic. A single overflowing
// line hands layout back to the terminal's naive wrap, which is what destroys
// the gutter and caret alignment.
func TestRenderNeverExceedsWidth(t *testing.T) {
	for name, d := range renderCorpus() {
		for width := diagWidthFloor; width <= 120; width++ {
			for _, line := range renderAt(d, width) {
				if w := com.DisplayWidth(line); w > width {
					t.Errorf("%s at width %d: line is %d columns:\n%s", name, width, w, line)
					break
				}
			}
		}
	}
}

// The caret has to point at the thing the message is about. Tabs and wide
// characters are where this has historically gone wrong, because spans arrive
// as byte offsets but carets are placed in display columns.
func TestCaretAlignsWithSpan(t *testing.T) {
	tests := []struct {
		name   string
		source string
		// byte offsets of the span within the (single-line) source
		startCol, endCol int
	}{
		{"ascii", "hello = 2 a", 10, 11},
		{"tab indented", "\t\tyes no", 6, 8},
		{"after cjk", `msg = "日本語" + broken`, 22, 28},
		{"leading", "x = 1", 0, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := diag(tt.source, "problem", 0, tt.startCol, tt.endCol)

			// Wide enough that no windowing kicks in, so the caret column is
			// directly comparable to the source column.
			lines := renderAt(d, 120)
			srcLine, caretLine := findSourceAndCaret(t, lines)

			expanded, colOf := com.ExpandTabs(tt.source)
			wantCol := colOf[tt.startCol]

			gutter := strings.Index(srcLine, "| ") + 2
			if got := com.DisplayWidth(srcLine[:gutter]) + wantCol; got != caretColumn(caretLine) {
				t.Errorf("caret at column %d, want %d\nsource:  %s\ncaret:   %s\nexpanded: %q",
					caretColumn(caretLine), got, srcLine, caretLine, expanded)
			}
		})
	}
}

// Trailing whitespace would force go-snap to store every diagnostic snapshot as
// a quoted string, which makes them unreadable to review.
func TestRenderHasNoTrailingWhitespace(t *testing.T) {
	for name, d := range renderCorpus() {
		for _, width := range []int{40, 60, 80, 100, 120} {
			for _, line := range renderAt(d, width) {
				if line != strings.TrimRight(line, " \t") {
					t.Errorf("%s at width %d: trailing whitespace in %q", name, width, line)
				}
			}
		}
	}
}

// A diagnostic that panics or silently drops its span is worse than a wide one.
func TestRenderAlwaysMarksTheSpan(t *testing.T) {
	for name, d := range renderCorpus() {
		if d.Source == "" {
			continue
		}
		for width := diagWidthFloor; width <= 120; width += 7 {
			out := strings.Join(renderAt(d, width), "\n")
			if !strings.Contains(out, "^") {
				t.Errorf("%s at width %d: no caret in output:\n%s", name, width, out)
			}
		}
	}
}

// Below the floor and above the cap, layout must stop changing - otherwise a
// wider terminal could produce narrower prose, which is the behavior we
// explicitly rejected.
func TestWidthClamping(t *testing.T) {
	for name, d := range renderCorpus() {
		narrow := strings.Join(renderAt(d, 10), "\n")
		floor := strings.Join(renderAt(d, diagWidthFloor), "\n")
		if narrow != floor {
			t.Errorf("%s: width 10 differs from the %d floor", name, diagWidthFloor)
		}

		// Above the cap only the source block may keep growing, so compare the
		// prose lines - those outside the "|" block.
		capped := proseLines(renderAt(d, diagWidthCap))
		wide := proseLines(renderAt(d, 400))
		if capped != wide {
			t.Errorf("%s: prose at width 400 differs from the %d cap:\n%s\n---\n%s",
				name, diagWidthCap, capped, wide)
		}
	}
}

func proseLines(lines []string) string {
	var out []string
	for _, l := range lines {
		if !strings.Contains(l, "|") {
			out = append(out, l)
		}
	}
	return strings.Join(out, "\n")
}

func findSourceAndCaret(t *testing.T, lines []string) (string, string) {
	t.Helper()
	for i, l := range lines {
		if i+1 < len(lines) && strings.Contains(lines[i+1], "^") {
			return l, lines[i+1]
		}
	}
	t.Fatalf("no source/caret pair in:\n%s", strings.Join(lines, "\n"))
	return "", ""
}

func caretColumn(caretLine string) int {
	return com.DisplayWidth(caretLine[:strings.Index(caretLine, "^")])
}
