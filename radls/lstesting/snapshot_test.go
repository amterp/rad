package lstesting

import (
	"os"
	"testing"

	"go.uber.org/zap"

	snap "github.com/amterp/go-snap"

	"github.com/amterp/rad/radls/log"
)

func TestMain(m *testing.M) {
	log.L = zap.NewNop().Sugar()
	os.Exit(m.Run())
}

func TestSnapshots(t *testing.T) {
	suite := Suite
	suite.Run = runCase
	snap.Run(t, "snapshots", &suite)
}

func runCase(t *testing.T, c *snap.Case) {
	tc, err := caseFrom(c)
	if err != nil {
		t.Fatalf("%v", err)
	}
	out, err := Run(tc)
	if err != nil {
		t.Fatalf("harness: %v", err)
	}
	c.Out("STDOUT", out)
}
