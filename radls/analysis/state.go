package analysis

import (
	"sync"

	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/rad/rts"
)

// State is the per-server registry of open documents. It owns the
// shared tree-sitter parser and the negotiated position encoding. The
// docs map is guarded by a top-level mutex for membership changes
// (open/close); per-document write coordination lives on Document
// itself, and reads of any document's content are lock-free via its
// snapshot pointer.
type State struct {
	parser   *rts.RadParser
	encoding PositionEncoding

	mu   sync.RWMutex
	docs map[string]*Document

	// parserMu serializes calls into the tree-sitter parser. Tree-sitter
	// parsers are NOT safe to share across goroutines; until/unless we
	// hand out per-document parsers, every Parse() call needs this lock.
	parserMu sync.Mutex
}

func NewState() *State {
	radParser, err := rts.NewRadParser()
	if err != nil {
		log.L.Fatalw("Failed to create Rad tree sitter", "err", err)
	}

	return &State{
		parser:   radParser,
		encoding: EncodingUTF16,
		docs:     make(map[string]*Document),
	}
}

// Encoding returns the LSP position encoding currently in use.
func (s *State) Encoding() PositionEncoding {
	return s.encoding
}

// SetEncoding installs the encoding negotiated at initialize. Must be
// called before any didOpen so the first document's diagnostics use
// the right encoding. Initialize happens exactly once per session, so
// we don't bother guarding against a second call.
func (s *State) SetEncoding(enc PositionEncoding) {
	s.encoding = enc
}

// Snapshot returns the current version of the named document, or nil
// if the document isn't open. Lock-free on the version side; the docs
// map lookup takes the RWMutex read lock briefly.
func (s *State) Snapshot(uri string) *DocumentVersion {
	s.mu.RLock()
	doc, ok := s.docs[uri]
	s.mu.RUnlock()
	if !ok {
		return nil
	}
	return doc.Snapshot()
}

// document returns the *Document handle, creating any missing entry is
// the caller's responsibility (do it through AddDoc).
func (s *State) document(uri string) *Document {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.docs[uri]
}

// AddDoc opens a document for the first time and produces its initial
// version. If the URI is already open the existing entry is replaced -
// LSP shouldn't send didOpen twice for the same URI, but be defensive.
func (s *State) AddDoc(uri, text string) {
	log.L.Infof("Adding doc %s", uri)
	doc := &Document{}
	doc.Update(func(_ *DocumentVersion) *DocumentVersion {
		return s.buildVersionLocked(uri, 1, text)
	})

	s.mu.Lock()
	s.docs[uri] = doc
	s.mu.Unlock()
}

// UpdateDoc applies a sequence of content changes and produces a fresh
// version for each. Today the protocol delivers full-document text
// (TextDocumentSync = 1), so each change spawns a wholesale reparse;
// the per-change loop is preserved so a future move to incremental
// edits is a one-spot change.
func (s *State) UpdateDoc(uri string, changes []lsp.TextDocumentContentChangeEvent) {
	doc := s.document(uri)
	if doc == nil {
		log.L.Warnw("UpdateDoc on unopened URI - ignoring", "uri", uri)
		return
	}
	for _, change := range changes {
		log.L.Infow("Updating doc", "uri", uri)
		doc.Update(func(prev *DocumentVersion) *DocumentVersion {
			var nextVer int64 = 1
			if prev != nil {
				nextVer = prev.version + 1
			}
			return s.buildVersionLocked(uri, nextVer, change.Text)
		})
	}
}

// buildVersionLocked builds a new DocumentVersion while holding
// parserMu, since tree-sitter parsers are not goroutine-safe.
func (s *State) buildVersionLocked(uri string, version int64, text string) *DocumentVersion {
	s.parserMu.Lock()
	defer s.parserMu.Unlock()
	return buildVersion(s.parser, s.encoding, uri, version, text)
}

// GetDiagnostics returns the diagnostics for the current version of
// the named document, or an empty slice if not open. Backwards
// compatibility shim - new code should grab a Snapshot directly.
func (s *State) GetDiagnostics(uri string) []lsp.Diagnostic {
	snap := s.Snapshot(uri)
	if snap == nil {
		return nil
	}
	return snap.Diagnostics()
}

// safeConvertCST converts a CST to AST, recovering from panics caused by
// invalid syntax during editing. Returns nil if conversion fails.
func safeConvertCST(tree *rts.RadTree, src, file string) (ast *rl.SourceFile) {
	defer func() {
		if r := recover(); r != nil {
			ast = nil
		}
	}()
	return rts.ConvertCST(tree.Root(), src, file)
}
