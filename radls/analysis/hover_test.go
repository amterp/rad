package analysis

import (
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// Behavior cases for hover live in
// radls/lstesting/snapshots/hover.snap, exercised through the
// real LSP wire harness. The Go test here only covers the
// nil-snapshot guard, which can't be reached through the wire
// (every wire request goes through a snapshot lookup that
// either finds a version or short-circuits before the
// analysis layer sees nil).

// TestHoverNoSnapshotReturnsNil verifies we don't crash when
// callers (today only the server handler, but the analysis
// method is public) pass a nil snapshot. The server's own
// nil-check should prevent this, but defensive is cheap.
func TestHoverNoSnapshotReturnsNil(t *testing.T) {
	s := NewState()
	h, err := s.Hover(nil, lsp.NewPos(0, 0))
	if err != nil {
		t.Fatalf("Hover: %v", err)
	}
	if h != nil {
		t.Errorf("expected nil hover for nil snapshot, got %+v", h)
	}
}
