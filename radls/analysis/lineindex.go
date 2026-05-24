package analysis

import (
	"unicode/utf8"
)

// PositionEncoding identifies how LSP positions are measured. LSP 3.17
// lets a client/server negotiate one of these at the initialize handshake;
// utf-16 is the default when no negotiation occurs.
type PositionEncoding string

const (
	EncodingUTF8  PositionEncoding = "utf-8"
	EncodingUTF16 PositionEncoding = "utf-16"
	EncodingUTF32 PositionEncoding = "utf-32"
)

// serverPositionEncodingPreference is the order radls prefers when picking
// from a client's offered list. utf-8 is cheapest because our internals
// (tree-sitter, source strings) are already utf-8 byte-indexed.
var serverPositionEncodingPreference = []PositionEncoding{
	EncodingUTF8,
	EncodingUTF16,
	EncodingUTF32,
}

// NegotiatePositionEncoding picks the server-preferred encoding from the
// list a client advertised in its initialize params. Per LSP 3.17, if the
// client offers nothing we must use utf-16. If the client offers something
// but we share nothing in common we also fall back to utf-16 (the spec's
// universal mandatory).
func NegotiatePositionEncoding(clientOffered []PositionEncoding) PositionEncoding {
	if len(clientOffered) == 0 {
		return EncodingUTF16
	}
	for _, want := range serverPositionEncodingPreference {
		for _, got := range clientOffered {
			if got == want {
				return want
			}
		}
	}
	return EncodingUTF16
}

// LineIndex translates positions on a fixed text snapshot. Tree-sitter
// reports positions in utf-8 byte columns; LSP clients want them in the
// encoding negotiated at initialize. LineIndex is the single conversion
// point so the rest of the analyzer can stay in byte units (matching
// tree-sitter) without ever caring about utf-16 surrogate pairs.
//
// Construction is O(n) over the text; per-position queries are O(byteCol)
// on the affected line because we have to walk the bytes to count code
// units. That's fine at our scale (Rad files are small and queries are
// per-diagnostic, not per-keystroke).
type LineIndex struct {
	text       string
	lineStarts []int
}

// NewLineIndex scans text once to record where each line begins.
func NewLineIndex(text string) *LineIndex {
	starts := make([]int, 1, 64)
	for i := 0; i < len(text); i++ {
		if text[i] == '\n' {
			starts = append(starts, i+1)
		}
	}
	return &LineIndex{text: text, lineStarts: starts}
}

// LineCount returns the number of lines in the indexed text. A document
// with no trailing newline still counts the partial final line.
func (l *LineIndex) LineCount() int {
	return len(l.lineStarts)
}

// lineSlice returns the byte content of line N excluding any trailing \n.
// Out-of-range lines return "".
func (l *LineIndex) lineSlice(line int) string {
	if line < 0 || line >= len(l.lineStarts) {
		return ""
	}
	start := l.lineStarts[line]
	var end int
	if line+1 < len(l.lineStarts) {
		end = l.lineStarts[line+1] - 1
	} else {
		end = len(l.text)
	}
	if end < start {
		end = start
	}
	return l.text[start:end]
}

// ByteColumnTo converts a utf-8 byte column on `line` into a column in
// the target encoding. Out-of-range or past-end-of-line inputs are
// clamped to the line's length in the target encoding so callers can't
// produce LSP positions that violate the spec.
func (l *LineIndex) ByteColumnTo(line, byteCol int, enc PositionEncoding) int {
	if byteCol < 0 {
		byteCol = 0
	}
	if enc == EncodingUTF8 {
		// Even for utf-8 we still want to clamp to line length.
		ls := l.lineSlice(line)
		if byteCol > len(ls) {
			return len(ls)
		}
		return byteCol
	}
	ls := l.lineSlice(line)
	if byteCol > len(ls) {
		byteCol = len(ls)
	}
	prefix := ls[:byteCol]
	switch enc {
	case EncodingUTF16:
		return utf16Units(prefix)
	case EncodingUTF32:
		return utf8.RuneCountInString(prefix)
	}
	return byteCol
}

// ColumnToByte is the inverse of ByteColumnTo: given a column reported in
// the client's encoding, return the utf-8 byte column tree-sitter uses
// internally. Out-of-range inputs clamp to the line's byte length so we
// never index past the end of a line.
func (l *LineIndex) ColumnToByte(line, col int, enc PositionEncoding) int {
	if col < 0 {
		col = 0
	}
	ls := l.lineSlice(line)
	if enc == EncodingUTF8 {
		if col > len(ls) {
			return len(ls)
		}
		return col
	}
	switch enc {
	case EncodingUTF16:
		return byteColFromUTF16(ls, col)
	case EncodingUTF32:
		return byteColFromUTF32(ls, col)
	}
	return col
}

// utf16Units returns how many utf-16 code units encode s. ASCII chars
// take one unit, basic-multilingual-plane non-ASCII still takes one,
// astral-plane code points (e.g. most emoji) take two via a surrogate
// pair. This mirrors what JavaScript's `s.length` would report - the
// historical reason LSP defaults to utf-16.
func utf16Units(s string) int {
	n := 0
	for _, r := range s {
		if r >= 0x10000 {
			n += 2
		} else {
			n++
		}
	}
	return n
}

// byteColFromUTF16 walks s and returns the byte index at which `target`
// utf-16 code units have been consumed. If target lands inside a
// surrogate pair we return the byte index of the start of that pair
// (clients shouldn't send mid-surrogate positions, but be defensive).
// If target exceeds the line, we clamp to len(s).
func byteColFromUTF16(s string, target int) int {
	if target <= 0 {
		return 0
	}
	consumed := 0
	for i, r := range s {
		if consumed >= target {
			return i
		}
		if r >= 0x10000 {
			consumed += 2
		} else {
			consumed++
		}
	}
	return len(s)
}

// byteColFromUTF32 returns the byte index at the Nth utf-32 code point.
// Equivalent to "byte index of the Nth rune."
func byteColFromUTF32(s string, target int) int {
	if target <= 0 {
		return 0
	}
	consumed := 0
	for i := range s {
		if consumed == target {
			return i
		}
		consumed++
	}
	return len(s)
}
