package analysis

import (
	"sync"
	"sync/atomic"

	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"
)

// DocumentVersion is an immutable snapshot of a single parse of a
// document. Multiple readers can hold the same *DocumentVersion
// concurrently without coordination: the data inside is guaranteed
// not to mutate. The next didChange produces a NEW DocumentVersion
// (built off the old) and atomically swaps the owning Document's
// current pointer.
//
// This is the load-bearing piece of Phase 8: it lets LSP request
// handlers (hover, goto-def, completion, etc.) grab a snapshot once
// and reason about a frozen world for the duration of the request,
// rather than racing against the next keystroke.
//
// Lifetime: the underlying tree-sitter tree owns C-heap memory that
// the Go GC can't reclaim. We refcount accordingly: the Document
// holds one reference; each call to State.Snapshot bumps the count
// and returns the snapshot to the caller, who MUST call Release()
// when done. When the count reaches zero, the tree is Close()d.
//
// Why refcounting and not a finalizer: tree-sitter Nodes carry
// references into the C tree that the Go GC doesn't see. A
// finalizer can fire on a tree while another goroutine is walking
// nodes that point into it - the documented cgo hazard. Explicit
// release gives us deterministic free at the moment we know no
// reader is touching the tree.
type DocumentVersion struct {
	uri         string
	fileID      FileID
	version     int64
	text        string
	tree        *rts.RadTree
	ast         *rl.SourceFile
	lineIndex   *LineIndex
	encoding    PositionEncoding
	diagnostics []lsp.Diagnostic
	// rawDiagnostics is the same data as `diagnostics` but in the
	// check.Diagnostic shape - utf-8 byte columns, full suggestion
	// strings, error codes. Code actions consume these to derive
	// quick-fix edits; the LSP-shape diagnostics have already lost
	// the suggestion and code fields by the time they reach the
	// wire, so re-deriving from the raw form is the right path.
	rawDiagnostics []check.Diagnostic
	// resolved and types are the analysis indexes the LSP feature
	// handlers (hover, goto-def, find-refs, completion) consult to
	// answer requests. They share a (frozen) AST with this snapshot,
	// so they're safe to read concurrently for the lifetime of the
	// version. Either may be nil if the source failed to convert
	// (e.g. mid-edit syntax error) - callers must handle that.
	resolved *check.Resolved
	types    *check.TypeInfo

	// refs starts at 1 - the Document that owns this version holds
	// the initial reference. Each State.Snapshot caller bumps to one
	// more; their matching Release drops it. When this hits zero the
	// underlying tree is freed and any later acquire() returns false.
	refs atomic.Int32
}

func (v *DocumentVersion) URI() string                   { return v.uri }
func (v *DocumentVersion) FileID() FileID                { return v.fileID }
func (v *DocumentVersion) Version() int64                { return v.version }
func (v *DocumentVersion) Text() string                  { return v.text }
func (v *DocumentVersion) Tree() *rts.RadTree            { return v.tree }
func (v *DocumentVersion) AST() *rl.SourceFile           { return v.ast }
func (v *DocumentVersion) LineIndex() *LineIndex         { return v.lineIndex }
func (v *DocumentVersion) Encoding() PositionEncoding    { return v.encoding }
func (v *DocumentVersion) Diagnostics() []lsp.Diagnostic { return v.diagnostics }
func (v *DocumentVersion) Resolved() *check.Resolved     { return v.resolved }
func (v *DocumentVersion) Types() *check.TypeInfo        { return v.types }

// acquire bumps the refcount if the snapshot is still live. Returns
// false if the snapshot has already been released (refs == 0), in
// which case the caller should retry the Document.Snapshot load -
// the State has a newer version.
//
// Uses CAS rather than a plain Add so we never resurrect a snapshot
// whose tree has already been Close()d. This is the standard
// "weak-to-strong reference upgrade" pattern.
func (v *DocumentVersion) acquire() bool {
	for {
		n := v.refs.Load()
		if n == 0 {
			return false
		}
		if v.refs.CompareAndSwap(n, n+1) {
			return true
		}
	}
}

// Release drops one reference. When the count reaches zero the
// underlying tree-sitter tree is freed. Each call to State.Snapshot
// pairs with exactly one Release - callers typically `defer
// snap.Release()` right after the nil-check.
func (v *DocumentVersion) Release() {
	if v == nil {
		return
	}
	if v.refs.Add(-1) == 0 {
		if v.tree != nil {
			v.tree.Close()
			v.tree = nil
		}
	}
}

// GetLine returns the source of the line at the given index, or "" if
// out of range. Kept on DocumentVersion (not LineIndex) because callers
// usually want both the text and the index together.
func (v *DocumentVersion) GetLine(line int) string {
	idx := v.lineIndex
	if line < 0 || line >= idx.LineCount() {
		return ""
	}
	return idx.lineSlice(line)
}

// Document owns the current snapshot of one LSP document. The snapshot
// pointer is read lock-free; writers serialize through `mu` to ensure
// a coherent prev->next chain (tree-sitter parsing today is wholesale,
// but a future incremental-parse path would still want this invariant).
//
// The FileID is fixed at construction; it stays stable across every
// version of the document so internal code can hold a FileID and
// always reach the latest snapshot via the State's lookup tables.
type Document struct {
	fileID   FileID
	snapshot atomic.Pointer[DocumentVersion]
	mu       sync.Mutex
}

func (d *Document) FileID() FileID { return d.fileID }

// Snapshot returns the current immutable version of this document.
// Lock-free; safe to call from any goroutine.
func (d *Document) Snapshot() *DocumentVersion {
	return d.snapshot.Load()
}

// Update runs `produce` under the writer lock to compute the next
// version from the previous (nil on first open), then atomically swaps
// it into place. The new version arrives with refs=1 (held by us);
// after the store we Release the previous version, dropping
// Document's reference to it. Any reader that had already Acquired
// the old version keeps it alive via the refcount.
func (d *Document) Update(produce func(prev *DocumentVersion) *DocumentVersion) *DocumentVersion {
	d.mu.Lock()
	defer d.mu.Unlock()
	prev := d.snapshot.Load()
	next := produce(prev)
	d.snapshot.Store(next)
	if prev != nil {
		prev.Release()
	}
	return next
}

// buildVersion is the canonical way to construct a DocumentVersion: it
// parses the source, builds the AST, indexes lines, runs the static
// checker, and translates diagnostics into the negotiated encoding.
// Caller (typically State) owns the parser and encoding it passes in.
func buildVersion(
	parser *rts.RadParser,
	encoding PositionEncoding,
	uri string,
	fileID FileID,
	version int64,
	text string,
) *DocumentVersion {
	tree := parser.Parse(text)
	ast := safeConvertCST(tree, text, uri)
	lineIndex := NewLineIndex(text)

	checker := check.NewCheckerWithTree(tree, parser, text, ast)
	diags, rawDiags, resolved, typeInfo := runChecker(checker, lineIndex, encoding)

	v := &DocumentVersion{
		uri:            uri,
		fileID:         fileID,
		version:        version,
		text:           text,
		tree:           tree,
		ast:            ast,
		lineIndex:      lineIndex,
		encoding:       encoding,
		diagnostics:    diags,
		rawDiagnostics: rawDiags,
		resolved:       resolved,
		types:          typeInfo,
	}
	// Owner's reference. Released by Document.Update when this
	// version is replaced by a successor.
	v.refs.Store(1)
	return v
}

// runChecker is the boundary between check.Diagnostic (utf-8 byte
// columns) and lsp.Diagnostic (negotiated encoding). Lives here so the
// snapshot construction path owns the translation, rather than scatter
// it across the analysis package.
//
// Returns the LSP-shaped diagnostics plus the analysis indexes from
// the same check pass. The indexes (resolved, typeInfo) are nil when
// AST conversion failed; they're load-bearing for LSP features beyond
// diagnostics (hover, goto-def, etc.).
func runChecker(checker check.RadChecker, idx *LineIndex, enc PositionEncoding) ([]lsp.Diagnostic, []check.Diagnostic, *check.Resolved, *check.TypeInfo) {
	out := make([]lsp.Diagnostic, 0)
	result, err := checker.Check()
	if err != nil {
		return out, nil, nil, nil
	}
	for _, cd := range result.Diagnostics {
		rang := lsp.Range{
			Start: lsp.Pos{
				Line:      cd.Range.Start.Line,
				Character: idx.ByteColumnTo(cd.Range.Start.Line, cd.Range.Start.Character, enc),
			},
			End: lsp.Pos{
				Line:      cd.Range.End.Line,
				Character: idx.ByteColumnTo(cd.Range.End.Line, cd.Range.End.Character, enc),
			},
		}
		out = append(out, lsp.NewDiagnosticFromCheckWithRange(cd, rang))
	}
	return out, result.Diagnostics, result.Resolved, result.Types
}
