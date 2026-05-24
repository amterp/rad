package analysis

import (
	"sync"
	"testing"

	"github.com/amterp/rad/radls/lsp"
)

// TestFileIDAssignment verifies a fresh FileID is allocated per
// document and the State maintains both URI->FileID and FileID->Document
// lookups correctly.
func TestFileIDAssignment(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)

	const uriA = "file:///a.rad"
	const uriB = "file:///b.rad"

	s.AddDoc(uriA, "x = 1")
	idA := s.FileIDFor(uriA)
	if idA == InvalidFileID {
		t.Fatal("expected valid FileID for opened doc")
	}

	s.AddDoc(uriB, "y = 2")
	idB := s.FileIDFor(uriB)
	if idB == InvalidFileID {
		t.Fatal("expected valid FileID for opened doc B")
	}
	if idA == idB {
		t.Errorf("expected distinct FileIDs, got %d == %d", idA, idB)
	}

	// FileIDFor on an unknown URI returns InvalidFileID.
	if got := s.FileIDFor("file:///never.rad"); got != InvalidFileID {
		t.Errorf("unknown URI: expected InvalidFileID, got %d", got)
	}

	// SnapshotByID round-trips correctly.
	snapA := s.SnapshotByID(idA)
	if snapA == nil {
		t.Fatal("SnapshotByID(idA) returned nil")
	}
	defer snapA.Release()
	if snapA.URI() != uriA {
		t.Errorf("SnapshotByID resolved to wrong doc: got URI %q, want %q",
			snapA.URI(), uriA)
	}
	if snapA.FileID() != idA {
		t.Errorf("snapshot FileID mismatch: got %d, want %d", snapA.FileID(), idA)
	}

	// SnapshotByID for an unknown id returns nil.
	if got := s.SnapshotByID(FileID(9999)); got != nil {
		t.Errorf("unknown id: expected nil, got %+v", got)
	}
}

// TestFileIDStableAcrossUpdates verifies the FileID stays constant
// when the document goes through versions. This is the property that
// makes FileID useful: internal code can pin a FileID and keep
// reaching the latest snapshot, no matter how many didChanges have
// flowed through.
func TestFileIDStableAcrossUpdates(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)

	const uri = "file:///stable.rad"
	s.AddDoc(uri, "v1")
	id1 := s.FileIDFor(uri)

	s.UpdateDoc(uri, []lsp.TextDocumentContentChangeEvent{{Text: "v2"}})
	id2 := s.FileIDFor(uri)
	if id1 != id2 {
		t.Errorf("FileID changed across versions: %d -> %d", id1, id2)
	}

	snap := s.SnapshotByID(id1)
	if snap == nil {
		t.Errorf("SnapshotByID after update: nil")
	} else {
		defer snap.Release()
		if snap.Text() != "v2" {
			t.Errorf("SnapshotByID after update: expected text v2, got %q", snap.Text())
		}
	}
}

// TestFileIDConcurrentAllocations confirms atomic allocation: many
// goroutines opening different docs all get distinct ids.
func TestFileIDConcurrentAllocations(t *testing.T) {
	s := NewState()
	s.SetEncoding(EncodingUTF16)

	const n = 100
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			uri := "file:///" + itoa(i) + ".rad"
			s.AddDoc(uri, "x = 1")
		}(i)
	}
	wg.Wait()

	seen := make(map[FileID]bool, n)
	for i := 0; i < n; i++ {
		uri := "file:///" + itoa(i) + ".rad"
		id := s.FileIDFor(uri)
		if id == InvalidFileID {
			t.Errorf("doc %s missing", uri)
			continue
		}
		if seen[id] {
			t.Errorf("duplicate id %d for %s", id, uri)
		}
		seen[id] = true
	}
}
