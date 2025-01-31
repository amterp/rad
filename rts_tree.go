package rts

import ts "github.com/tree-sitter/go-tree-sitter"

type RtsTree struct {
	root   *ts.Tree
	parser *ts.Parser
	src    string
}

func NewRtsTree(tree *ts.Tree, parser *ts.Parser, src string) *RtsTree {
	return &RtsTree{
		root:   tree,
		parser: parser,
		src:    src,
	}
}

func (rt *RtsTree) Update(src string) {
	// todo use incremental parsing, maybe can lean on LSP client to give via protocol
	rt.root = rt.parser.Parse([]byte(src), nil)
}

func (rt *RtsTree) Close() {
	rt.root.Close()
}

func (rt *RtsTree) String() string {
	return rt.root.RootNode().ToSexp()
}

func (rt *RtsTree) GetShebang() (*Shebang, bool) {
	node, ok := rt.findNode("shebang", rt.root.RootNode())
	if ok {
		return &Shebang{
			BaseNode: BaseNode{
				Src:       rt.src[node.StartByte():node.EndByte()],
				StartByte: int(node.StartByte()),
				EndByte:   int(node.EndByte()),
				StartPos:  NewPosition(node.StartPosition()),
				EndPos:    NewPosition(node.EndPosition()),
			}}, true
	}
	return nil, false
}

func (rt *RtsTree) GetFileHeader() (*FileHeader, bool) {
	node, ok := rt.findNode("file_header", rt.root.RootNode())
	if !ok {
		return nil, false
	}

	contentsNode := node.ChildByFieldName("contents")
	if contentsNode == nil {
		// would be strange`
		return nil, false
	}

	return &FileHeader{
		BaseNode: BaseNode{
			Src:       rt.src[node.StartByte():node.EndByte()],
			StartByte: int(node.StartByte()),
			EndByte:   int(node.EndByte()),
			StartPos:  NewPosition(node.StartPosition()),
			EndPos:    NewPosition(node.EndPosition()),
		},
		Contents: rt.src[contentsNode.StartByte():contentsNode.EndByte()],
	}, true
}

func (rt *RtsTree) findNode(nodeKind string, node *ts.Node) (*ts.Node, bool) {
	if node.Kind() == nodeKind {
		return node, true
	}
	children := node.Children(node.Walk())
	for _, child := range children {
		if n, ok := rt.findNode(nodeKind, &child); ok {
			return n, true
		}
	}
	return nil, false
}
