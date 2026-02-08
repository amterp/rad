package analysis

import (
	"strings"

	"github.com/amterp/rad/lsp-server/log"
	"github.com/amterp/rad/lsp-server/lsp"

	"github.com/amterp/rad/rts/check"
	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/rad/rts"
)

type DocState struct {
	uri         string
	text        string
	tree        *rts.RadTree
	ast         *rl.SourceFile
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

type State struct {
	parser *rts.RadParser
	// URI -> Text
	docs map[string]*DocState
}

func NewState() *State {
	radParser, err := rts.NewRadParser()
	if err != nil {
		log.L.Fatalw("Failed to create Rad tree sitter", "err", err)
	}

	return &State{
		parser: radParser,
		docs:   make(map[string]*DocState),
	}
}

func (s *State) NewDocState(uri, text string) *DocState {
	tree := s.parser.Parse(text)
	ast := safeConvertCST(tree, text, uri)
	checker := check.NewCheckerWithTree(tree, s.parser, text, ast)
	return &DocState{
		uri:         uri,
		text:        text,
		tree:        tree,
		ast:         ast,
		diagnostics: s.resolveDiagnostics(checker),
		checker:     checker,
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
		doc.ast = safeConvertCST(doc.tree, change.Text, doc.uri)
		doc.checker.Update(doc.tree, change.Text, doc.ast)
		doc.diagnostics = s.resolveDiagnostics(doc.checker)
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
