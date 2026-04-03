package main

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"
)

// TestHomebrewStartupMessage mirrors the radls test in the homebrew-core
// formula (Formula/r/rad.rb). If this test breaks, the homebrew formula
// test will also break. Update both together.
//
// The formula asserts:
//
//	assert_match "Spinning up Rad LSP server", shell_output("#{bin}/radls 2>&1", 1)
func TestHomebrewStartupMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "go", "run", ".")
	cmd.Stdin = strings.NewReader("")
	output, _ := cmd.CombinedOutput()

	// The homebrew formula expects this exact string in stderr.
	expected := "Spinning up Rad LSP server"
	if !strings.Contains(string(output), expected) {
		t.Errorf("Expected output to contain %q, got:\n%s", expected, output)
	}
}
