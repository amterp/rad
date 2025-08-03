package analysis

import (
	"strings"

	"github.com/amterp/rad/lsp-server/com"
	"github.com/amterp/rad/lsp-server/log"
	"github.com/amterp/rad/lsp-server/lsp"

	"github.com/amterp/rad/rts/check"

	"github.com/amterp/rad/rts"
)

type DocState struct {
	uri         string
	text        string
	tree        *rts.RadTree
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
	parser       *rts.RadParser
	radFunctions *com.FunctionSet
	// URI -> Text
	docs map[string]*DocState
}

func NewState() *State {
	radParser, err := rts.NewRadParser()
	if err != nil {
		log.L.Fatalw("Failed to create Rad tree sitter", "err", err)
	}
	radFunctions := com.LoadNewFunctionSet()
	log.L.Infof("Loaded %d functions", radFunctions.Len())

	return &State{
		parser:       radParser,
		radFunctions: radFunctions,
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
