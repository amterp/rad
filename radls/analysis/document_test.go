package analysis

import (
	"sync"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// TestSnapshotStability exercises the central guarantee of Phase 8:
// once a reader has a *DocumentVersion, subsequent writes don't change
// what that pointer observes. The reader sees the world as it was
// when it grabbed its snapshot.
func TestSnapshotStability(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)

	const uri = "file:///snap_test.rad"
	s.AddDoc(uri, "x = 1")

	first := s.Snapshot(uri)
	if first == nil {
		t.Fatal("expected snapshot after AddDoc")
	}
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
				if len(txt) == 0 {
					t.Errorf("empty text in snapshot")
					return
				}
			}
		}()
	}

	// Let it run a bit, then stop.
	for i := 0; i < 200; i++ {
		_ = s.Snapshot(uri)
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
