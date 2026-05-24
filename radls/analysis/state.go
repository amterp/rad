package analysis

import (
	"strings"

	"github.com/amterp/rad/radls/log"
	"github.com/amterp/rad/radls/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/rad/rts"
)

type DocState struct {
	uri         string
	text        string
	tree        *rts.RadTree
	ast         *rl.SourceFile
	lineIndex   *LineIndex
	diagnostics []lsp.Diagnostic
	checker     check.RadChecker
}

func (d *DocState) GetLine(line int) string {
	// todo wasteful implementation
	lines := strings.Split(d.text, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	return lines[line]
}

// LineIndex returns the index over this document's current text.
// Always non-nil for live documents - safe to call after AddDoc.
func (d *DocState) LineIndex() *LineIndex {
	return d.lineIndex
}

type State struct {
	parser *rts.RadParser
	// encoding is the LSP position encoding negotiated at initialize.
	// Defaults to utf-16 (the LSP-mandatory baseline) until SetEncoding
	// is called by the server during handleInitialize.
	encoding PositionEncoding
	// URI -> Text
	docs map[string]*DocState
}

func NewState() *State {
	radParser, err := rts.NewRadParser()
	if err != nil {
		log.L.Fatalw("Failed to create Rad tree sitter", "err", err)
	}

	return &State{
		parser:   radParser,
		encoding: EncodingUTF16,
		docs:     make(map[string]*DocState),
	}
}

// Encoding returns the LSP position encoding the server is currently
// using to talk to its client.
func (s *State) Encoding() PositionEncoding {
	return s.encoding
}

// SetEncoding installs the encoding negotiated at initialize. Must be
// called before any didOpen so that diagnostics on the first document
// use the right encoding. Calling it twice would be a protocol violation
// (initialize happens exactly once); we just overwrite anyway since the
// blast radius is small and any sane client will only call us once.
func (s *State) SetEncoding(enc PositionEncoding) {
	s.encoding = enc
}

func (s *State) NewDocState(uri, text string) *DocState {
	tree := s.parser.Parse(text)
	ast := safeConvertCST(tree, text, uri)
	checker := check.NewCheckerWithTree(tree, s.parser, text, ast)
	doc := &DocState{
		uri:       uri,
		text:      text,
		tree:      tree,
		ast:       ast,
		lineIndex: NewLineIndex(text),
		checker:   checker,
	}
	doc.diagnostics = s.resolveDiagnostics(doc)
	return doc
}

func (s *State) AddDoc(uri, text string) {
	log.L.Infof("Adding doc %s", uri)
	s.docs[uri] = s.NewDocState(uri, text)
}

func (s *State) UpdateDoc(uri string, changes []lsp.TextDocumentContentChangeEvent) {
	doc := s.docs[uri]
	for _, change := range changes {
		log.L.Infow("Updating doc", "uri", uri)
		log.L.Debugf("Tree before: %s", doc.tree.String())
		doc.tree.Update(change.Text)
		log.L.Debugf("Tree after: %s", doc.tree.String())
		doc.text = change.Text
		doc.ast = safeConvertCST(doc.tree, change.Text, doc.uri)
		doc.lineIndex = NewLineIndex(change.Text)
		doc.checker.Update(doc.tree, change.Text, doc.ast)
		doc.diagnostics = s.resolveDiagnostics(doc)
	}
}

func (s *State) GetDiagnostics(uri string) []lsp.Diagnostic {
	return s.docs[uri].diagnostics
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
