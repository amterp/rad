package analysis

import (
	"rls/com"
	"rls/log"
	"rls/lsp"
	"strings"

	"github.com/amterp/rts/check"

	"github.com/amterp/rts"
)

type DocState struct {
	uri         string
	text        string
	tree        *rts.RslTree
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
	parser       *rts.RslParser
	rslFunctions *com.FunctionSet
	// URI -> Text
	docs map[string]*DocState
}

func NewState() *State {
	rslParser, err := rts.NewRslParser()
	if err != nil {
		log.L.Fatalw("Failed to create RSL tree sitter", "err", err)
	}
	rslFunctions := com.LoadNewFunctionSet()
	log.L.Infof("Loaded %d functions", rslFunctions.Len())

	return &State{
		parser:       rslParser,
		rslFunctions: rslFunctions,
		docs:         make(map[string]*DocState),
	}
}

func (s *State) NewDocState(uri, text string) *DocState {
	tree := s.parser.Parse(text)
	checker := check.NewCheckerWithTree(tree, s.parser, text)
	return &DocState{
		uri:         uri,
		text:        text,
		tree:        tree,
		diagnostics: s.resolveDiagnostics(tree, checker),
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
		doc.checker.UpdateSrc(change.Text) // todo CHECKER WILL REPEAT PARSE, BAD
		doc.diagnostics = s.resolveDiagnostics(doc.tree, doc.checker)
	}
}

func (s *State) GetDiagnostics(uri string) []lsp.Diagnostic {
	return s.docs[uri].diagnostics
}
