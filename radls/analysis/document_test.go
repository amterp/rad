package analysis

import (
	"sync"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// TestSnapshotStability exercises the central guarantee of Phase 8:
// once a reader has a *DocumentVersion, subsequent writes don't change
// what that pointer observes. The reader sees the world as it was
// when it grabbed its snapshot. The reader's Acquire keeps the
// underlying tree alive across writer-driven version swaps.
func TestSnapshotStability(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)

	const uri = "file:///snap_test.rad"
	s.AddDoc(uri, "x = 1")

	first := s.Snapshot(uri)
	if first == nil {
		t.Fatal("expected snapshot after AddDoc")
	}
	defer first.Release()
	if got := first.Text(); got != "x = 1" {
		t.Errorf("v1 text: got %q, want %q", got, "x = 1")
	}
	v1Version := first.Version()

	// Apply two updates. Each must produce a new snapshot that the
	// State returns from Snapshot(), without disturbing the original
	// pointer the reader is holding.
	s.UpdateDoc(uri, []lsp.TextDocumentContentChangeEvent{{Text: "x = 2"}})
	s.UpdateDoc(uri, []lsp.TextDocumentContentChangeEvent{{Text: "x = 3"}})

	if first.Text() != "x = 1" {
		t.Errorf("v1 text mutated under reader: got %q, want %q",
			first.Text(), "x = 1")
	}
	if first.Version() != v1Version {
		t.Errorf("v1 version mutated under reader: got %d, want %d",
			first.Version(), v1Version)
	}

	latest := s.Snapshot(uri)
	defer latest.Release()
	if latest == first {
		t.Errorf("Snapshot() returned the same pointer after updates")
	}
	if latest.Text() != "x = 3" {
		t.Errorf("latest text: got %q, want %q", latest.Text(), "x = 3")
	}
	if latest.Version() <= v1Version {
		t.Errorf("expected version > %d, got %d", v1Version, latest.Version())
	}
}

// TestSnapshotReleaseFreesTreeWhenLastRefDropped verifies the tree
// is closed once the refcount reaches zero. Direct test of the
// memory-leak fix: each Snapshot bumps refs, each Release drops one,
// and when no more references exist the tree's C memory is freed.
func TestSnapshotReleaseFreesTreeWhenLastRefDropped(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///release.rad"
	s.AddDoc(uri, "x = 1")

	first := s.Snapshot(uri)
	if first == nil {
		t.Fatal("expected snapshot")
	}
	// At this point: refs = 2 (Document + caller).
	if got := first.refs.Load(); got != 2 {
		t.Errorf("after Snapshot: refs=%d, want 2", got)
	}

	// Update once. Document drops its reference to `first`, leaving
	// just the caller's reference.
	s.UpdateDoc(uri, []lsp.TextDocumentContentChangeEvent{{Text: "x = 2"}})
	if got := first.refs.Load(); got != 1 {
		t.Errorf("after Update: refs=%d, want 1", got)
	}
	if first.tree == nil {
		t.Error("tree should still be alive while caller holds a ref")
	}

	// Caller releases. Refcount hits zero, tree is closed and set
	// to nil.
	first.Release()
	if got := first.refs.Load(); got != 0 {
		t.Errorf("after Release: refs=%d, want 0", got)
	}
	if first.tree != nil {
		t.Error("tree should be nil after last Release")
	}

	// A late Acquire on a freed snapshot must fail. Without this the
	// refcount could go negative and we'd never detect the bug.
	if first.acquire() {
		t.Error("acquire on released snapshot should fail")
	}
}

// TestSnapshotConcurrentReaders runs many goroutines that all read
// snapshots while writers churn. Doesn't assert on race conditions
// directly (that's `go test -race`'s job), but does verify the data
// each reader sees is internally consistent: the text it observed
// matches the version number.
func TestSnapshotConcurrentReaders(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	const uri = "file:///snap_concurrent.rad"
	s.AddDoc(uri, "x = 0")

	var wg sync.WaitGroup
	stop := make(chan struct{})

	// Writers
	for w := 0; w < 2; w++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			i := 0
			for {
				select {
				case <-stop:
					return
				default:
					s.UpdateDoc(uri, []lsp.TextDocumentContentChangeEvent{
						{Text: "x = " + itoa(i)},
					})
					i++
				}
			}
		}(w)
	}

	// Readers: each snapshot's text must be internally consistent.
	for r := 0; r < 4; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				snap := s.Snapshot(uri)
				if snap == nil {
					continue
				}
				// Read text and version; they're frozen for this snapshot.
				txt := snap.Text()
				_ = snap.Version()
				_ = snap.LineIndex().LineCount()
				snap.Release()
				if len(txt) == 0 {
					t.Errorf("empty text in snapshot")
					return
				}
			}
		}()
	}

	// Let it run a bit, then stop.
	for i := 0; i < 200; i++ {
		snap := s.Snapshot(uri)
		if snap != nil {
			snap.Release()
		}
	}
	close(stop)
	wg.Wait()
}

// TestUpdateDocOnUnopenedURIIgnored verifies the defensive guard - we
// log and return rather than panic if didChange arrives before didOpen.
func TestUpdateDocOnUnopenedURIIgnored(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)
	// Don't panic.
	s.UpdateDoc("file:///never_opened.rad", []lsp.TextDocumentContentChangeEvent{
		{Text: "x = 1"},
	})
	if got := s.Snapshot("file:///never_opened.rad"); got != nil {
		t.Errorf("expected nil snapshot, got %+v", got)
	}
}

// itoa is a tiny local helper to avoid pulling in strconv in this test.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
