package docir

import (
	"strings"
	"unicode/utf8"
)

// emitTable renders a Table as a column-aligned GFM table. Aligning at
// emit time means the result reads correctly whether it's printed raw
// (piped) or run through the runtime terminal renderer - the renderer
// never needs to understand tables. Pipes are kept so the output is
// still a valid table for anything that does parse GFM (e.g. an LLM
// reading the piped corpus).
func emitTable(t Table) string {
	cols := len(t.Header)
	if cols == 0 {
		return ""
	}

	widths := make([]int, cols)
	measure := func(cells []string) {
		for i := 0; i < cols && i < len(cells); i++ {
			if w := utf8.RuneCountInString(cells[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}
	measure(t.Header)
	for _, r := range t.Rows {
		measure(r)
	}
	for i := range widths {
		if widths[i] < 3 { // room for the separator's dashes
			widths[i] = 3
		}
	}

	var b strings.Builder
	b.WriteString(renderRow(t.Header, widths, t.Align))
	b.WriteString(renderSeparator(widths, t.Align))
	for _, r := range t.Rows {
		b.WriteString(renderRow(r, widths, t.Align))
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderRow(cells []string, widths []int, align []Align) string {
	var b strings.Builder
	b.WriteString("|")
	for i, w := range widths {
		cell := ""
		if i < len(cells) {
			cell = cells[i]
		}
		b.WriteString(" ")
		b.WriteString(pad(cell, w, alignAt(align, i)))
		b.WriteString(" |")
	}
	b.WriteString("\n")
	return b.String()
}

func renderSeparator(widths []int, align []Align) string {
	var b strings.Builder
	b.WriteString("|")
	for i, w := range widths {
		b.WriteString(" ")
		b.WriteString(dashes(w, alignAt(align, i)))
		b.WriteString(" |")
	}
	b.WriteString("\n")
	return b.String()
}

func dashes(width int, align Align) string {
	switch align {
	case AlignLeft:
		return ":" + strings.Repeat("-", width-1)
	case AlignRight:
		return strings.Repeat("-", width-1) + ":"
	case AlignCenter:
		return ":" + strings.Repeat("-", width-2) + ":"
	default:
		return strings.Repeat("-", width)
	}
}

func pad(s string, width int, align Align) string {
	gap := width - utf8.RuneCountInString(s)
	if gap <= 0 {
		return s
	}
	switch align {
	case AlignRight:
		return strings.Repeat(" ", gap) + s
	case AlignCenter:
		left := gap / 2
		return strings.Repeat(" ", left) + s + strings.Repeat(" ", gap-left)
	default:
		return s + strings.Repeat(" ", gap)
	}
}

func alignAt(align []Align, i int) Align {
	if i < len(align) {
		return align[i]
	}
	return AlignNone
}
