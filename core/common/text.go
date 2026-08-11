package com

import (
	"strings"

	tblwriter "github.com/amterp/go-tbl"
	"github.com/mattn/go-runewidth"
)

// TabWidth is how many columns a tab occupies when a tab-indented source line
// is rendered in a diagnostic. Rad's own formatter indents with spaces, so this
// only applies to hand-written source that uses tabs.
const TabWidth = 4

// DisplayWidth reports how many terminal columns a string occupies, ignoring
// ANSI escape sequences and accounting for wide (CJK) and zero-width runes.
//
// Distinct from StrLen, which counts runes: a CJK character is one rune but two
// columns. Anything laying text out against a terminal width wants this one.
func DisplayWidth(s string) int {
	return tblwriter.DisplayWidth(s)
}

// Wrap greedily fills lines of at most width columns, breaking on whitespace.
// A word wider than width on its own is hard-broken rather than allowed to
// overflow, so every returned line satisfies DisplayWidth(line) <= width.
// Always returns at least one line.
//
// The single exception is a lone rune wider than width - a CJK character at
// width 1. It is emitted overflowing rather than dropped, since the alternative
// is silently losing text or looping forever. Callers wrapping to any sane
// terminal width never reach it.
//
// Wrap takes PLAIN text. Colorize the lines it returns, never the string you
// pass in: a break landing inside an escape sequence would split it, and even a
// clean break leaves the sequence unterminated on the line before. Wrapping
// first and coloring after also keeps line breaks identical with and without
// --color, which is what lets a single snapshot cover both.
//
// Greedy rather than balanced (go-tbl's WrapString minimizes raggedness) for
// two reasons: greedy is what terminal diagnostics conventionally look like,
// and it re-flows locally. Adding a word to a balanced paragraph moves every
// line, which turns every message edit into an unreviewable snapshot diff.
func Wrap(s string, width int) []string {
	return wrapLines(s, func(int) int { return width })
}

// WrapPrefixed wraps s to width, prefixing the first line with first and the
// rest with cont. Each line is filled against its own prefix, so a wide "= help: "
// tag on line one doesn't shorten every line after it.
func WrapPrefixed(s, first, cont string, width int) []string {
	firstW, contW := DisplayWidth(first), DisplayWidth(cont)

	lines := wrapLines(s, func(i int) int {
		if i == 0 {
			return width - firstW
		}
		return width - contW
	})

	for i := range lines {
		if i == 0 {
			lines[i] = first + lines[i]
		} else {
			lines[i] = cont + lines[i]
		}
	}
	return lines
}

// wrapLines is the greedy fill behind Wrap and WrapPrefixed. widthAt gives the
// budget for line i, consulted afresh on every break so callers can vary it by
// line - which is what makes a hanging indent fill its lines properly.
func wrapLines(s string, widthAt func(int) int) []string {
	var lines []string
	var line strings.Builder
	lineWidth := 0

	budget := func(i int) int {
		return IntMax(1, widthAt(i))
	}
	width := budget(0)

	flush := func() {
		lines = append(lines, line.String())
		line.Reset()
		lineWidth = 0
		width = budget(len(lines))
	}

	for _, word := range strings.Fields(s) {
		wordWidth := DisplayWidth(word)

		if lineWidth > 0 && lineWidth+1+wordWidth > width {
			flush()
		}

		// A word too wide to ever fit gets chopped across as many lines as it
		// needs. Rare in prose, but file paths and URLs hit it on narrow
		// terminals, and overflowing would defeat the whole exercise. The line
		// is always empty here: anything wider than the full width also failed
		// the fit test above, so it was already flushed.
		for wordWidth > width {
			var head string
			head, word = splitAtWidth(word, width)
			line.WriteString(head)
			flush()
			wordWidth = DisplayWidth(word)
		}

		if lineWidth > 0 {
			line.WriteByte(' ')
			lineWidth++
		}
		line.WriteString(word)
		lineWidth += wordWidth
	}

	if lineWidth > 0 || len(lines) == 0 {
		flush()
	}
	return lines
}

// ExpandTabs replaces tabs with spaces to the next TabWidth boundary and
// returns the expanded text alongside a byte-offset -> display-column table
// with len(line)+1 entries.
//
// The table is the point: tree-sitter reports spans as byte offsets, but a
// caret has to land at a display column. Without it, tabs push the caret out of
// alignment and any wide character shifts everything after it.
func ExpandTabs(line string) (string, []int) {
	var sb strings.Builder
	colOf := make([]int, 0, len(line)+1)

	col := 0
	for _, r := range line {
		// One entry per byte of the rune, so a byte offset landing mid-rune
		// still maps to that rune's starting column.
		for i := 0; i < len(string(r)); i++ {
			colOf = append(colOf, col)
		}

		if r == '\t' {
			width := TabWidth - col%TabWidth
			sb.WriteString(strings.Repeat(" ", width))
			col += width
			continue
		}

		sb.WriteRune(r)
		col += runewidth.RuneWidth(r)
	}
	colOf = append(colOf, col)

	return sb.String(), colOf
}

// SliceColumns returns the part of s between display columns [start, end).
// A wide rune straddling either boundary is dropped rather than half-drawn, so
// the result never exceeds end-start columns.
func SliceColumns(s string, start, end int) string {
	var sb strings.Builder
	col := 0
	for _, r := range s {
		w := runewidth.RuneWidth(r)
		if col >= start && col+w <= end {
			sb.WriteRune(r)
		}
		col += w
		if col >= end {
			break
		}
	}
	return sb.String()
}

// ShortenPathLeft trims a path from the left to at most width columns, marking
// the cut with "...". Paths lose their least useful information at the front,
// so "/Users/me/src/proj/main.rad" shortens to ".../proj/main.rad".
func ShortenPathLeft(path string, width int) string {
	pathWidth := DisplayWidth(path)
	if pathWidth <= width || width < 4 {
		return path
	}
	return "..." + SliceColumns(path, pathWidth-(width-3), pathWidth)
}

// splitAtWidth splits s into the longest prefix fitting in width columns and
// the remainder. Always consumes at least one rune when s is non-empty, even if
// that rune is wider than width - otherwise a 2-column character at width 1
// would make no progress and hang the caller.
func splitAtWidth(s string, width int) (string, string) {
	col := 0
	for i, r := range s {
		w := runewidth.RuneWidth(r)
		if i > 0 && col+w > width {
			return s[:i], s[i:]
		}
		col += w
	}
	return s, ""
}
