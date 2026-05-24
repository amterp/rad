package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// Behavior cases for findReferences live in
// radls/lstesting/snapshots/references.snap. This file
// keeps only the nil-snapshot guard.

func TestReferencesNoSnapshotReturnsEmpty(t *testing.T) {
	s := NewState()
	locs, err := s.References(nil, lsp.NewPos(0, 0), true)
	if err != nil {
		t.Fatalf("References: %v", err)
	}
	if locs == nil || len(locs) != 0 {
		t.Errorf("expected empty slice for nil snapshot, got %v", locs)
	}
}
