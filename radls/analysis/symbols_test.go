package analysis

import (
	"testing"
)

// Behavior cases for documentSymbol live in
// radls/lstesting/snapshots/document_symbol.snap. This file
// keeps only the nil-snapshot guard.

func TestDocumentSymbolsNoSnapshotReturnsEmpty(t *testing.T) {
	s := NewState()
	syms, err := s.DocumentSymbols(nil)
	if err != nil {
		t.Fatalf("DocumentSymbols: %v", err)
	}
	if syms == nil {
		t.Errorf("expected empty slice for nil snapshot, got nil")
	}
	if len(syms) != 0 {
		t.Errorf("expected 0 symbols for nil snapshot, got %d", len(syms))
	}
}
