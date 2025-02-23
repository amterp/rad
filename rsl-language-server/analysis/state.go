package analysis

import (
	"rls/log"
	"rls/lsp"
	"strings"

	"github.com/amterp/rts"
)

type DocState struct {
	uri         string
	text        string
	tree        *rts.RslTree
	diagnostics []lsp.Diagnostic
}

func (d *DocState) GetLine(line int) string {
	// todo wasteful implementation
	lines := strings.Split(d.text, "\n")
	if line < 0 || line >= len(lines) {
		return ""
	}
	return lines[line]
}

type State struct {
	parser *rts.RslParser
	// URI -> Text
	docs map[string]*DocState
}

func NewState() *State {
	rslParser, err := rts.NewRslParser()
	if err != nil {
		log.L.Fatalw("Failed to create RSL tree sitter", "err", err)
	}

	return &State{
		parser: rslParser,
		docs:   make(map[string]*DocState),
	}
}

func (s *State) NewDocState(uri, text string) *DocState {
	tree := s.parser.Parse(text)
	return &DocState{
		uri:         uri,
		text:        text,
		tree:        tree,
		diagnostics: resolveDiagnostics(tree),
	}
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
		doc.diagnostics = resolveDiagnostics(doc.tree)
	}
}

func (s *State) GetDiagnostics(uri string) []lsp.Diagnostic {
	return s.docs[uri].diagnostics
}

// todo be able to give yet better diagnostics e.g. unknown functions, etc. Where should live though?
func resolveDiagnostics(tree *rts.RslTree) []lsp.Diagnostic {
	invalidNodes := tree.FindInvalidNodes()
	diagnostics := make([]lsp.Diagnostic, len(invalidNodes))
	for i, node := range invalidNodes {
		rang := lsp.NewRangeFromTsNode(node)
		diagnostics[i] = lsp.NewDiagnostic(rang, lsp.Err, "RSL Language Server", "Invalid syntax")
	}
	return diagnostics
}
