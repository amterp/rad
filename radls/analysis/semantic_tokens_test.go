package analysis

import (
	"testing"
)

// Behavior cases for semanticTokens live in
// radls/lstesting/snapshots/semantic_tokens.snap. This file
// keeps two invariant checks that don't fit a wire-level
// snapshot:
//
//  - LegendStable: the wire-encoding index/name pairing has to
//    match the in-Go TokenType constants. A drift here would
//    silently mis-color tokens. Pure-Go assertion is the right
//    shape; a snapshot would just dump the same legend without
//    pinning the invariant.
//  - NoSnapshotReturnsEmpty: the nil-snapshot guard, unreachable
//    through the wire harness.

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
