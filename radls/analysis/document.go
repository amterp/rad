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
// current pointer; old versions remain valid for any reader still
// holding them, and are GC'd once unreferenced.
//
// This is the load-bearing piece of Phase 8: it lets LSP request
// handlers (hover, goto-def, completion, etc.) grab a snapshot once
// and reason about a frozen world for the duration of the request,
// rather than racing against the next keystroke.
type DocumentVersion struct {
	uri         string
	version     int64
	text        string
	tree        *rts.RadTree
	ast         *rl.SourceFile
	lineIndex   *LineIndex
	diagnostics []lsp.Diagnostic
}

func (v *DocumentVersion) URI() string                  { return v.uri }
func (v *DocumentVersion) Version() int64               { return v.version }
func (v *DocumentVersion) Text() string                 { return v.text }
func (v *DocumentVersion) Tree() *rts.RadTree           { return v.tree }
func (v *DocumentVersion) AST() *rl.SourceFile          { return v.ast }
func (v *DocumentVersion) LineIndex() *LineIndex        { return v.lineIndex }
func (v *DocumentVersion) Diagnostics() []lsp.Diagnostic { return v.diagnostics }

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
type Document struct {
	snapshot atomic.Pointer[DocumentVersion]
	mu       sync.Mutex
}

// Snapshot returns the current immutable version of this document.
// Lock-free; safe to call from any goroutine.
func (d *Document) Snapshot() *DocumentVersion {
	return d.snapshot.Load()
}

// Update runs `produce` under the writer lock to compute the next
// version from the previous (nil on first open), then atomically swaps
// it into place. Returns the new version.
func (d *Document) Update(produce func(prev *DocumentVersion) *DocumentVersion) *DocumentVersion {
	d.mu.Lock()
	defer d.mu.Unlock()
	prev := d.snapshot.Load()
	next := produce(prev)
	d.snapshot.Store(next)
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
	version int64,
	text string,
) *DocumentVersion {
	tree := parser.Parse(text)
	ast := safeConvertCST(tree, text, uri)
	lineIndex := NewLineIndex(text)

	checker := check.NewCheckerWithTree(tree, parser, text, ast)
	diags := runChecker(checker, lineIndex, encoding)

	return &DocumentVersion{
		uri:         uri,
		version:     version,
		text:        text,
		tree:        tree,
		ast:         ast,
		lineIndex:   lineIndex,
		diagnostics: diags,
	}
}

// runChecker is the boundary between check.Diagnostic (utf-8 byte
// columns) and lsp.Diagnostic (negotiated encoding). Lives here so the
// snapshot construction path owns the translation, rather than scatter
// it across the analysis package.
func runChecker(checker check.RadChecker, idx *LineIndex, enc PositionEncoding) []lsp.Diagnostic {
	out := make([]lsp.Diagnostic, 0)
	result, err := checker.Check()
	if err != nil {
		return out
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
	return out
}
