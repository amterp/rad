package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// Behavior cases for goto-definition live in
// radls/lstesting/snapshots/definition.snap, exercised through
// the LSP wire harness. This file keeps only the nil-snapshot
// guard - the wire never reaches the analysis layer with a nil
// snapshot, but the public method is callable directly so we
// keep it covered.

func TestDefinitionNoSnapshotReturnsNil(t *testing.T) {
	s := NewState()
	loc, err := s.Definition(nil, lsp.NewPos(0, 0))
	if err != nil {
		t.Fatalf("Definition: %v", err)
	}
	if loc != nil {
		t.Errorf("expected nil location for nil snapshot, got %+v", loc)
	}
}
