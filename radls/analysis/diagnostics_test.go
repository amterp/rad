package analysis

import (
	"testing"

	"github.com/amterp/rad/rts/check"
)

// TestToLspRangeUTF16 verifies the byte->utf-16 boundary conversion.
// Documents the contract: check.Range carries utf-8 byte columns (the
// tree-sitter native); resolveDiagnostics translates to the client's
// negotiated encoding before publishing.
func TestToLspRangeUTF16(t *testing.T) {
	// Line 0: `x = "中"`  (9 bytes; 中 = 3 bytes, 1 utf-16 unit)
	// Line 1: `y = "🎉"` (10 bytes; 🎉 = 4 bytes, 2 utf-16 units)
	text := "x = \"中\"\ny = \"🎉\""

	s := NewState()
	s.SetEncoding(EncodingUTF16)
	doc := &DocState{lineIndex: NewLineIndex(text)}

	cases := []struct {
		name     string
		in       check.Range
		wantSL   int
		wantSC   int
		wantEL   int
		wantEC   int
	}{
		{
			name:   "ascii range untouched",
			in:     mkRange(0, 0, 0, 3),
			wantSL: 0, wantSC: 0, wantEL: 0, wantEC: 3,
		},
		{
			name:   "byte col after CJK char compresses to utf-16",
			in:     mkRange(0, 4, 0, 9), // covers `"中"` (5 bytes -> 3 utf-16 units)
			wantSL: 0, wantSC: 4, wantEL: 0, wantEC: 7,
		},
		{
			name:   "byte col after astral char compresses",
			in:     mkRange(1, 4, 1, 10), // covers `"🎉"` (6 bytes -> 4 utf-16 units)
			wantSL: 1, wantSC: 4, wantEL: 1, wantEC: 8,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := s.toLspRange(tc.in, doc)
			if got.Start.Line != tc.wantSL || got.Start.Character != tc.wantSC ||
				got.End.Line != tc.wantEL || got.End.Character != tc.wantEC {
				t.Errorf("got %d:%d-%d:%d, want %d:%d-%d:%d",
					got.Start.Line, got.Start.Character, got.End.Line, got.End.Character,
					tc.wantSL, tc.wantSC, tc.wantEL, tc.wantEC)
			}
		})
	}
}

// TestToLspRangeUTF8 verifies utf-8 is a passthrough (with clamping).
// A client that negotiates utf-8 should get byte columns unchanged.
func TestToLspRangeUTF8(t *testing.T) {
	text := "x = \"中\""

	s := NewState()
	s.SetEncoding(EncodingUTF8)
	doc := &DocState{lineIndex: NewLineIndex(text)}

	in := mkRange(0, 4, 0, 9)
	got := s.toLspRange(in, doc)
	if got.Start.Character != 4 || got.End.Character != 9 {
		t.Errorf("utf-8 passthrough: got %d-%d, want 4-9",
			got.Start.Character, got.End.Character)
	}
}

func mkRange(sl, sc, el, ec int) check.Range {
	return check.Range{
		Start: check.Pos{Line: sl, Character: sc},
		End:   check.Pos{Line: el, Character: ec},
	}
}
