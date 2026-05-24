package analysis

import "testing"

func TestNegotiatePositionEncoding(t *testing.T) {
	cases := []struct {
		name    string
		offered []PositionEncoding
		want    PositionEncoding
	}{
		{"no offer falls back to utf-16", nil, EncodingUTF16},
		{"empty offer falls back to utf-16", []PositionEncoding{}, EncodingUTF16},
		{"utf-8 alone is taken", []PositionEncoding{EncodingUTF8}, EncodingUTF8},
		{"utf-16 alone is taken", []PositionEncoding{EncodingUTF16}, EncodingUTF16},
		{"utf-32 alone is taken", []PositionEncoding{EncodingUTF32}, EncodingUTF32},
		{"utf-8 wins over utf-16",
			[]PositionEncoding{EncodingUTF16, EncodingUTF8}, EncodingUTF8},
		{"utf-8 wins over utf-32",
			[]PositionEncoding{EncodingUTF32, EncodingUTF8}, EncodingUTF8},
		{"utf-16 wins over utf-32 when no utf-8",
			[]PositionEncoding{EncodingUTF32, EncodingUTF16}, EncodingUTF16},
		{"unknown values fall back to utf-16",
			[]PositionEncoding{"utf-7"}, EncodingUTF16},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := NegotiatePositionEncoding(tc.offered)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLineIndexLineCount(t *testing.T) {
	cases := []struct {
		text string
		want int
	}{
		{"", 1},
		{"abc", 1},
		{"abc\n", 2},
		{"a\nb\nc", 3},
		{"a\nb\nc\n", 4},
	}
	for _, tc := range cases {
		t.Run(tc.text, func(t *testing.T) {
			got := NewLineIndex(tc.text).LineCount()
			if got != tc.want {
				t.Errorf("text %q: got %d lines, want %d", tc.text, got, tc.want)
			}
		})
	}
}

func TestLineIndexByteColumnTo(t *testing.T) {
	// Layout:
	//   line 0: x = 1
	//   line 1: y = "é"      (é = c3 a9 in utf-8, 1 unit in utf-16/32)
	//   line 2: z = "中"     (中 = e4 b8 ad in utf-8, 1 unit in utf-16/32)
	//   line 3: e = "🎉"     (🎉 = f0 9f 8e 89 in utf-8, 2 units in utf-16, 1 in utf-32)
	text := "x = 1\n" +
		"y = \"é\"\n" +
		"z = \"中\"\n" +
		"e = \"🎉\""
	idx := NewLineIndex(text)

	// On the é line, the closing quote sits at byte column 7 (after 4-byte
	// prefix `y = `, opening quote at 4, é spans bytes 5-6, closing quote
	// at 7). In utf-16 / utf-32 the closing quote is at column 6.
	if got := idx.ByteColumnTo(1, 7, EncodingUTF8); got != 7 {
		t.Errorf("utf-8 closing quote on é line: got %d, want 7", got)
	}
	if got := idx.ByteColumnTo(1, 7, EncodingUTF16); got != 6 {
		t.Errorf("utf-16 closing quote on é line: got %d, want 6", got)
	}
	if got := idx.ByteColumnTo(1, 7, EncodingUTF32); got != 6 {
		t.Errorf("utf-32 closing quote on é line: got %d, want 6", got)
	}

	// On the 中 line, closing quote sits at byte col 8 (中 takes 3 bytes),
	// utf-16 / utf-32 col 6.
	if got := idx.ByteColumnTo(2, 8, EncodingUTF16); got != 6 {
		t.Errorf("utf-16 closing quote on 中 line: got %d, want 6", got)
	}

	// Line 3 is `e = "🎉"` - 10 bytes total. Closing quote at byte col 9.
	// In utf-16 that's col 7 (e + space + = + space + " = 5, surrogate
	// pair for 🎉 = 2). In utf-32 that's col 6 (one rune for 🎉). End of
	// line at byte col 10 = utf-16 col 8 = utf-32 col 7.
	if got := idx.ByteColumnTo(3, 9, EncodingUTF16); got != 7 {
		t.Errorf("utf-16 closing quote on 🎉 line: got %d, want 7", got)
	}
	if got := idx.ByteColumnTo(3, 9, EncodingUTF32); got != 6 {
		t.Errorf("utf-32 closing quote on 🎉 line: got %d, want 6", got)
	}
	if got := idx.ByteColumnTo(3, 10, EncodingUTF16); got != 8 {
		t.Errorf("utf-16 end of 🎉 line: got %d, want 8", got)
	}
	if got := idx.ByteColumnTo(3, 10, EncodingUTF32); got != 7 {
		t.Errorf("utf-32 end of 🎉 line: got %d, want 7", got)
	}

	// Clamping: past end of line clamps to line length.
	if got := idx.ByteColumnTo(0, 999, EncodingUTF16); got != 5 {
		t.Errorf("clamp past EOL: got %d, want 5", got)
	}
	// Out-of-range line is treated as empty; col clamps to 0.
	if got := idx.ByteColumnTo(99, 4, EncodingUTF16); got != 0 {
		t.Errorf("out-of-range line: got %d, want 0", got)
	}
	// Negative col clamps to 0.
	if got := idx.ByteColumnTo(0, -3, EncodingUTF16); got != 0 {
		t.Errorf("negative col: got %d, want 0", got)
	}
}

func TestLineIndexColumnToByte(t *testing.T) {
	// Same fixture as above.
	text := "x = 1\n" +
		"y = \"é\"\n" +
		"z = \"中\"\n" +
		"e = \"🎉\""
	idx := NewLineIndex(text)

	// On é line: utf-16 col 6 is the closing quote, which is byte col 7.
	if got := idx.ColumnToByte(1, 6, EncodingUTF16); got != 7 {
		t.Errorf("é line utf-16 col 6: got %d, want 7", got)
	}
	// On 🎉 line: utf-16 col 7 sits at the closing quote, byte col 9.
	if got := idx.ColumnToByte(3, 7, EncodingUTF16); got != 9 {
		t.Errorf("🎉 line utf-16 col 7: got %d, want 9", got)
	}
	// utf-32 col 6 on 🎉 line is also the closing quote, byte col 9.
	if got := idx.ColumnToByte(3, 6, EncodingUTF32); got != 9 {
		t.Errorf("🎉 line utf-32 col 6: got %d, want 9", got)
	}
	// Clamp past end of line.
	if got := idx.ColumnToByte(0, 999, EncodingUTF16); got != 5 {
		t.Errorf("clamp past EOL utf-16: got %d, want 5", got)
	}
	// utf-8 passes through and clamps.
	if got := idx.ColumnToByte(0, 3, EncodingUTF8); got != 3 {
		t.Errorf("utf-8 passthrough: got %d, want 3", got)
	}
	if got := idx.ColumnToByte(0, 999, EncodingUTF8); got != 5 {
		t.Errorf("utf-8 clamp: got %d, want 5", got)
	}
}

func TestLineIndexRoundTrip(t *testing.T) {
	// ByteColumnTo and ColumnToByte should be inverses across the line.
	lines := []string{
		"hello world",
		"é + é = ée",
		"中文测试",
		"mix 🎉 and 中 in one",
		"",
	}
	encodings := []PositionEncoding{EncodingUTF8, EncodingUTF16, EncodingUTF32}

	for _, line := range lines {
		idx := NewLineIndex(line)
		for byteCol := 0; byteCol <= len(line); byteCol++ {
			// We only test grapheme-boundary byte cols (start of a rune)
			// because non-boundary inputs aren't meaningful to LSP clients.
			if byteCol > 0 && byteCol < len(line) {
				b := line[byteCol]
				if b >= 0x80 && b < 0xC0 {
					continue // utf-8 continuation byte
				}
			}
			for _, enc := range encodings {
				col := idx.ByteColumnTo(0, byteCol, enc)
				back := idx.ColumnToByte(0, col, enc)
				if back != byteCol {
					t.Errorf("line=%q enc=%s byteCol=%d -> col=%d -> back=%d",
						line, enc, byteCol, col, back)
				}
			}
		}
	}
}

func TestLineIndexEmptyText(t *testing.T) {
	idx := NewLineIndex("")
	if got := idx.ByteColumnTo(0, 0, EncodingUTF16); got != 0 {
		t.Errorf("empty text byteColumnTo: got %d, want 0", got)
	}
	if got := idx.ColumnToByte(0, 0, EncodingUTF16); got != 0 {
		t.Errorf("empty text columnToByte: got %d, want 0", got)
	}
	if got := idx.LineCount(); got != 1 {
		t.Errorf("empty text line count: got %d, want 1", got)
	}
}
