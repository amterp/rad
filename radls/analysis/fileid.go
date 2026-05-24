package analysis

import "sync/atomic"

// FileID is an opaque identifier the analyzer uses to refer to one
// document. It's separate from the LSP URI on purpose:
//
//   - URIs are strings, expensive to hash and pass around in deep
//     traversals (symbol lookups, cross-file references, etc.)
//   - URIs are externally controlled - any uniqueness invariant lives
//     outside our type system, so the compiler can't help us when we
//     accidentally hand the wrong string to a function expecting a
//     "real" document handle
//   - URIs survive forever in client memory but a server-side close
//     should be able to recycle/invalidate references; an opaque int
//     ID gives us the seam to do that without changing the wire
//
// FileIDs are minted monotonically per State and never re-used within
// a session. Comparing FileIDs is O(1) and they fit in a register.
// The State maintains both URI -> FileID and FileID -> *Document
// maps so we can translate at the wire boundary and keep everything
// downstream typed.
type FileID uint32

// InvalidFileID is the sentinel returned when a lookup misses. Real
// FileIDs are always >= 1 - we start the counter at 1 so zero stays
// meaningful as "no file" / "not assigned yet."
const InvalidFileID FileID = 0

// fileIDAllocator hands out monotonically increasing FileIDs. Lives
// on the State; uses atomic so future concurrent didOpens (when the
// mux dispatches in goroutines) don't need a separate lock.
type fileIDAllocator struct {
	next atomic.Uint32
}

func (a *fileIDAllocator) Next() FileID {
	// Add(1) returns the post-increment value; first id is 1.
	return FileID(a.next.Add(1))
}
