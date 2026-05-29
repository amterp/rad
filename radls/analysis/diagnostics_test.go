package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// TestRunCheckerConvertsUTF16 verifies the byte->utf-16 boundary
// conversion that runChecker applies when translating check
// diagnostics into the wire format. Documents the contract:
// check.Range carries utf-8 byte columns (tree-sitter native);
// buildVersion translates to the client's negotiated encoding.
func TestRunCheckerConvertsUTF16(t *testing.T) {
	// Line 0: `x = "中"`  (9 bytes; 中 = 3 bytes, 1 utf-16 unit)
	// Line 1: `y = "🎉"` (10 bytes; 🎉 = 4 bytes, 2 utf-16 units)
	text := "x = \"中\"\ny = \"🎉\""
	idx := NewLineIndex(text)

	cases := []struct {
		name   string
		in     check.Range
		enc    PositionEncoding
		wantSC int
		wantEC int
	}{
		{"ascii range untouched utf-16", mkRange(0, 0, 0, 3), EncodingUTF16, 0, 3},
		{"byte col after CJK -> utf-16", mkRange(0, 4, 0, 9), EncodingUTF16, 4, 7},
		{"byte col after astral -> utf-16", mkRange(1, 4, 1, 10), EncodingUTF16, 4, 8},
		{"utf-8 passthrough", mkRange(0, 4, 0, 9), EncodingUTF8, 4, 9},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ck := &stubChecker{result: check.Result{Diagnostics: []check.Diagnostic{
				{Range: tc.in, Severity: check.Error, Message: "x"},
			}}}
			diags, _, _, _ := runChecker(ck, idx, tc.enc)
			if len(diags) != 1 {
				t.Fatalf("expected 1 diagnostic, got %d", len(diags))
			}
			got := diags[0].Range
			if got.Start.Character != tc.wantSC || got.End.Character != tc.wantEC {
				t.Errorf("got %d-%d, want %d-%d",
					got.Start.Character, got.End.Character,
					tc.wantSC, tc.wantEC)
			}
		})
	}
}

// TestRunCheckerEmpty verifies the no-diagnostics path still returns
// a non-nil (zero-length) slice, since the LSP wire format wants an
// explicit [] for "no diagnostics."
func TestRunCheckerEmpty(t *testing.T) {
	ck := &stubChecker{result: check.Result{Diagnostics: nil}}
	got, _, _, _ := runChecker(ck, NewLineIndex(""), EncodingUTF16)
	if got == nil {
		t.Errorf("expected non-nil empty slice, got nil")
	}
	if len(got) != 0 {
		t.Errorf("expected zero diagnostics, got %d", len(got))
	}
}

func mkRange(sl, sc, el, ec int) check.Range {
	return check.Range{
		Start: check.Pos{Line: sl, Character: sc},
		End:   check.Pos{Line: el, Character: ec},
	}
}

// stubChecker is a minimal RadChecker for testing the conversion path
// without needing a real parser/tree.
type stubChecker struct {
	result check.Result
}

func (s *stubChecker) UpdateSrc(string)                            {}
func (s *stubChecker) Update(*rts.RadTree, string, *rl.SourceFile) {}
func (s *stubChecker) Check() (check.Result, error)                { return s.result, nil }

// Compile-time interface check. lsp import keeps the file honest if
// the test ever drops its only lsp reference.
var _ check.RadChecker = (*stubChecker)(nil)
var _ = lsp.Diagnostic{}
