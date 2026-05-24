package rts

import (
	"sync"
	"testing"
)

// TestRadTreeCloseIdempotent verifies double-Close is safe. Upstream
// ts.Tree.Close is NOT idempotent (it always calls ts_tree_delete);
// our sync.Once guard protects against double-free.
func TestRadTreeCloseIdempotent(t *testing.T) {
	parser, err := NewRadParser()
	if err != nil {
		t.Fatalf("NewRadParser: %v", err)
	}
	defer parser.Close()

	tree := parser.Parse("x = 1")
	tree.Close()
	tree.Close()
	tree.Close()
}

// TestRadTreeUpdateSwapsRoot verifies Update actually installs a
// new underlying ts.Tree (not just mutates the old).
func TestRadTreeUpdateSwapsRoot(t *testing.T) {
	parser, err := NewRadParser()
	if err != nil {
		t.Fatalf("NewRadParser: %v", err)
	}
	defer parser.Close()

	rt := parser.Parse("x = 1")
	defer rt.Close()
	originalRoot := rt.root
	rt.Update("y = 2")
	if rt.root == originalRoot {
		t.Error("Update should replace the underlying tree")
	}
}

// TestRadTreeConcurrentClose verifies many goroutines calling Close
// at once is safe. The closeOnce should serialize.
func TestRadTreeConcurrentClose(t *testing.T) {
	parser, err := NewRadParser()
	if err != nil {
		t.Fatalf("NewRadParser: %v", err)
	}
	defer parser.Close()

	rt := parser.Parse("x = 1")
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rt.Close()
		}()
	}
	wg.Wait()
}
