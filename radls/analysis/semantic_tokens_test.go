package analysis

import (
	"testing"
)

func semanticTokensFixture(t *testing.T, src string) []uint {
	t.Helper()
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///sem_test.rad"
	s.AddDoc(uri, src)
	snap := s.Snapshot(uri)
	if snap == nil {
		t.Fatal("expected snapshot")
	}
	defer snap.Release()
	out, err := s.SemanticTokens(snap)
	if err != nil {
		t.Fatalf("SemanticTokens: %v", err)
	}
	return out.Data
}

// TestSemanticTokensEmptyDocument verifies an empty doc returns
// an empty (not nil) data slice. LSP wire wants an array.
func TestSemanticTokensEmptyDocument(t *testing.T) {
	data := semanticTokensFixture(t, "")
	if data == nil {
		t.Errorf("expected empty slice, got nil")
	}
	if len(data) != 0 {
		t.Errorf("expected 0 tokens, got %d (%v)", len(data), data)
	}
}

// TestSemanticTokensSingleLine verifies tokens on one line are
// emitted as (deltaLine, deltaStartChar, length, type, modifiers).
// For `print(x)`, we expect two tokens: print (Function) and x
// (Variable, after declaration on a previous line).
func TestSemanticTokensSingleLine(t *testing.T) {
	src := "x = 1\nprint(x)\n"
	data := semanticTokensFixture(t, src)
	// Expected tokens (in order):
	//   line 0 col 0: x  (Variable, length 1)
	//   line 1 col 0: print (Function, length 5)
	//   line 1 col 6: x (Variable, length 1)
	// Each token = 5 uints, so total = 15.
	if len(data) != 15 {
		t.Fatalf("expected 15 uints (3 tokens), got %d (%v)", len(data), data)
	}
	// First token: deltaLine=0, deltaStartChar=0, length=1, type=Variable, modifier=0
	if data[0] != 0 || data[1] != 0 || data[2] != 1 ||
		data[3] != uint(TokenTypeVariable) || data[4] != 0 {
		t.Errorf("first token: got %v, want (0,0,1,Variable,0)", data[0:5])
	}
	// Second token: deltaLine=1 (1-0), deltaStartChar=0 (absolute on new line), length=5, type=Function
	if data[5] != 1 || data[6] != 0 || data[7] != 5 ||
		data[8] != uint(TokenTypeFunction) || data[9] != 0 {
		t.Errorf("second token: got %v, want (1,0,5,Function,0)", data[5:10])
	}
	// Third token: deltaLine=0, deltaStartChar=6 (relative to print at col 0), length=1, Variable
	if data[10] != 0 || data[11] != 6 || data[12] != 1 ||
		data[13] != uint(TokenTypeVariable) || data[14] != 0 {
		t.Errorf("third token: got %v, want (0,6,1,Variable,0)", data[10:15])
	}
}

// TestSemanticTokensFnNameAtDefSite verifies the fn name in
// `fn greet():` is tagged Function. The binder declares hoisted
// fns at the FnDef node (not at an Identifier), so the name
// token has no Uses entry; we synthesize it from FnDef.NameSpan
// directly. Before this fix, only call sites of user fns were
// coloured - the decl site rendered as plain text.
func TestSemanticTokensFnNameAtDefSite(t *testing.T) {
	src := "fn greet():\n    print(1)\n"
	data := semanticTokensFixture(t, src)
	if len(data) == 0 {
		t.Fatal("expected tokens")
	}
	// First token should be `greet` on line 0 starting at col 3
	// (after `fn `). Function-kind, length 5.
	if data[0] != 0 || data[1] != 3 || data[2] != 5 ||
		data[3] != uint(TokenTypeFunction) || data[4] != 0 {
		t.Errorf("first token (greet at fn def): got %v, want (0,3,5,Function,0)",
			data[0:5])
	}
}

// TestSemanticTokensParamTagged verifies parameters of an enclosing
// function are tagged Parameter, not just Variable. This is the
// usability win - editors render params distinctly.
func TestSemanticTokensParamTagged(t *testing.T) {
	src := "fn greet(who: str):\n    print(who)\n"
	data := semanticTokensFixture(t, src)
	if len(data) == 0 {
		t.Fatal("expected non-empty token data")
	}
	// Find a Parameter-typed token among the data. The encoding is
	// (delta, delta, length, type, mod) per 5-uint group.
	foundParam := false
	for i := 0; i+4 < len(data); i += 5 {
		if data[i+3] == uint(TokenTypeParameter) {
			foundParam = true
			break
		}
	}
	if !foundParam {
		t.Errorf("expected at least one Parameter-typed token, data=%v",
			data)
	}
}

// TestSemanticTokensLegendStable verifies the legend's indices
// match the constants. The wire encoding refers to types by
// index, so any drift between SemanticTokensLegend() and the
// TokenType constants would produce silently-mistyped tokens.
func TestSemanticTokensLegendStable(t *testing.T) {
	legend := SemanticTokensLegend()
	if legend.TokenTypes[TokenTypeFunction] != "function" {
		t.Errorf("Function index: %q, want function", legend.TokenTypes[TokenTypeFunction])
	}
	if legend.TokenTypes[TokenTypeParameter] != "parameter" {
		t.Errorf("Parameter index: %q, want parameter", legend.TokenTypes[TokenTypeParameter])
	}
	if legend.TokenTypes[TokenTypeVariable] != "variable" {
		t.Errorf("Variable index: %q, want variable", legend.TokenTypes[TokenTypeVariable])
	}
}

// TestSemanticTokensNoSnapshotReturnsEmpty verifies nil-snapshot path.
func TestSemanticTokensNoSnapshotReturnsEmpty(t *testing.T) {
	s := NewState()
	out, err := s.SemanticTokens(nil)
	if err != nil {
		t.Fatalf("SemanticTokens: %v", err)
	}
	if out == nil {
		t.Fatal("expected non-nil SemanticTokens result")
	}
	if out.Data == nil {
		t.Errorf("expected empty data slice, got nil")
	}
}

// TestSemanticTokensUnresolvedIdentsSkipped verifies that an
// identifier that didn't resolve through the binder doesn't get
// a token. We don't want to color typos as if they were valid
// variables.
func TestSemanticTokensUnresolvedIdentsSkipped(t *testing.T) {
	// `undeclared_name` doesn't exist; it's read-but-never-written.
	// The binder won't put it in Uses.
	src := "print(undeclared_name)\n"
	data := semanticTokensFixture(t, src)
	// We expect ONE token: print. The undeclared name is skipped.
	if len(data) != 5 {
		t.Errorf("expected exactly one token (print), got %d uints (%v)",
			len(data), data)
	}
}
