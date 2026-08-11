package com

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestDisplayWidth(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"ascii", "hello", 5},
		{"empty", "", 0},
		{"cjk is double width", "日本語", 6},
		{"mixed", "a日b", 4},
		{"ansi is invisible", "\x1b[31mred\x1b[0m", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := DisplayWidth(tt.in); got != tt.want {
				t.Errorf("DisplayWidth(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestWrap(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		width int
		want  []string
	}{
		{"fits", "hello world", 20, []string{"hello world"}},
		{"exact fit", "hello world", 11, []string{"hello world"}},
		{"one over", "hello world", 10, []string{"hello", "world"}},
		{"greedy fill", "a b c d e f", 5, []string{"a b c", "d e f"}},
		{"collapses whitespace", "a  \n b", 10, []string{"a b"}},
		{"empty yields one line", "", 10, []string{""}},
		{"hard break", "abcdefghij", 4, []string{"abcd", "efgh", "ij"}},
		{"hard break after a word", "hi abcdefghij", 4, []string{"hi", "abcd", "efgh", "ij"}},
		{"cjk hard break", "日本語", 4, []string{"日本", "語"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Wrap(tt.in, tt.width)
			if len(got) != len(tt.want) {
				t.Fatalf("Wrap(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Wrap(%q, %d) line %d = %q, want %q", tt.in, tt.width, i, got[i], tt.want[i])
				}
			}
		})
	}
}

// The invariant everything else rests on: no line Wrap returns may overflow the
// width it was given. A diagnostic that overflows is the bug this all exists to
// fix, so it is worth asserting across the whole awkward corner of the space -
// widths narrower than a single character, and text that cannot be broken on
// whitespace.
func TestWrapNeverOverflows(t *testing.T) {
	inputs := []string{
		"",
		"a",
		"short words only here",
		"supercalifragilisticexpialidocious",
		"https://amterp.dev/rad/migrations/v0.12/ is the link",
		"日本語のテキストです",
		"mixed 日本語 and ascii together",
		strings.Repeat("x", 500),
		"\x1b[31mcolored\x1b[0m text",
	}
	for _, in := range inputs {
		for width := 1; width <= 40; width++ {
			for _, line := range Wrap(in, width) {
				w := DisplayWidth(line)
				if w <= width {
					continue
				}
				// The one documented exception: a single rune too wide for the
				// limit is emitted rather than dropped or looped on.
				if utf8.RuneCountInString(line) == 1 {
					continue
				}
				t.Errorf("Wrap(%q, %d) produced %q (width %d)", in, width, line, w)
			}
		}
	}
}

func TestWrapPrefixed(t *testing.T) {
	got := WrapPrefixed("reorder the targets to match", "= help: ", "        ", 20)
	want := []string{
		"= help: reorder the",
		"        targets to",
		"        match",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}

func TestExpandTabs(t *testing.T) {
	tests := []struct {
		name     string
		in       string
		expanded string
		// colOf entries to spot-check, as byteOffset -> displayColumn.
		cols map[int]int
	}{
		{"no tabs", "abc", "abc", map[int]int{0: 0, 1: 1, 3: 3}},
		{"leading tab", "\tx", "    x", map[int]int{0: 0, 1: 4, 2: 5}},
		{"tab mid-line", "ab\tc", "ab  c", map[int]int{2: 2, 3: 4}},
		{"tab at boundary", "abcd\te", "abcd    e", map[int]int{4: 4, 5: 8}},
		{"two tabs", "\t\tx", "        x", map[int]int{1: 4, 2: 8}},
		// The byte offsets of "日" are 0,1,2 - all mapping to column 0, so a
		// span reported mid-rune still points at the rune's own column.
		{"cjk", "日x", "日x", map[int]int{0: 0, 1: 0, 2: 0, 3: 2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expanded, colOf := ExpandTabs(tt.in)
			if expanded != tt.expanded {
				t.Errorf("ExpandTabs(%q) = %q, want %q", tt.in, expanded, tt.expanded)
			}
			if len(colOf) != len(tt.in)+1 {
				t.Fatalf("colOf has %d entries, want %d", len(colOf), len(tt.in)+1)
			}
			for offset, want := range tt.cols {
				if colOf[offset] != want {
					t.Errorf("colOf[%d] = %d, want %d", offset, colOf[offset], want)
				}
			}
		})
	}
}

func TestSliceColumns(t *testing.T) {
	tests := []struct {
		name       string
		in         string
		start, end int
		want       string
	}{
		{"whole", "abcdef", 0, 6, "abcdef"},
		{"middle", "abcdef", 2, 4, "cd"},
		{"past end", "abcdef", 4, 100, "ef"},
		{"empty range", "abcdef", 3, 3, ""},
		{"cjk aligned", "日本語", 2, 4, "本"},
		// Slicing into the middle of a wide rune drops it rather than drawing
		// half of it - a half-character would push everything after it out of
		// alignment, which is exactly what the caret math cannot tolerate.
		{"cjk straddling start", "日本語", 1, 4, "本"},
		{"cjk straddling end", "日本語", 0, 3, "日"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SliceColumns(tt.in, tt.start, tt.end); got != tt.want {
				t.Errorf("SliceColumns(%q, %d, %d) = %q, want %q", tt.in, tt.start, tt.end, got, tt.want)
			}
		})
	}
}

func TestShortenPathLeft(t *testing.T) {
	tests := []struct {
		name  string
		in    string
		width int
		want  string
	}{
		{"fits", "src/main.rad", 20, "src/main.rad"},
		{"exact", "src/main.rad", 12, "src/main.rad"},
		{"trims from the left", "/Users/me/src/proj/main.rad", 15, "...roj/main.rad"},
		// Too narrow to fit even the marker plus a character, so shortening
		// would communicate nothing. Hand the path back and let it overflow.
		{"width too small to mark", "src/main.rad", 3, "src/main.rad"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShortenPathLeft(tt.in, tt.width); got != tt.want {
				t.Errorf("ShortenPathLeft(%q, %d) = %q, want %q", tt.in, tt.width, got, tt.want)
			}
		})
	}
}
