package analysis

import (
	"os"
	"testing"

	"go.uber.org/zap"

	"github.com/amterp/rad/radls/log"
)

// TestMain wires up a no-op logger so tests that touch State / Document
// (which call log.L.Infof) don't nil-panic. We don't want test output
// to actually log anywhere, so a Nop is right.
func TestMain(m *testing.M) {
	log.L = zap.NewNop().Sugar()
	os.Exit(m.Run())
}
