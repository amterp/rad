package rts

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

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

func (rt *RtsTree) Sexp() string {
	return rt.root.RootNode().ToSexp()
}

func (rt *RtsTree) String() string {
	return rt.Dump()
}

func (rt *RtsTree) GetShebang() (*Shebang, bool) {
	node, ok := rt.findNode("shebang", rt.root.RootNode())
	if !ok {
		return nil, false
	}
	return newShebang(rt.src, node)
}

func (rt *RtsTree) GetFileHeader() (*FileHeader, bool) {
	node, ok := rt.findNode("file_header", rt.root.RootNode())
	if !ok {
		return nil, false
	}
	return newFileHeader(rt.src, node)
}

func QueryNodes[T Node](rt *RtsTree) ([]T, error) {
	nodeName := NodeName[T]()
	query, err := ts.NewQuery(rt.parser.Language(), fmt.Sprintf("(%s) @%s", nodeName, nodeName))

	if err != nil {
		return nil, err
	}

	qc := ts.NewQueryCursor()
	defer qc.Close()

	captures := qc.Captures(query, rt.root.RootNode(), nil)

	var nodes []T
	for {
		next, _ := captures.Next()
		if next == nil {
			break
		}

		node, ok := createNode[T](rt.src, &next.Captures[0].Node)
		if ok {
			nodes = append(nodes, node)
		}
	}
	return nodes, nil
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

func createNode[T Node](src string, node *ts.Node) (T, bool) {
	var zero T
	switch any(zero).(type) {
	case *Shebang:
		shebang, _ := newShebang(src, node)
		return any(shebang).(T), true
	case *FileHeader:
		fileHeader, _ := newFileHeader(src, node)
		return any(fileHeader).(T), true
	case *StringNode:
		stringNode, _ := newStringNode(src, node)
		return any(stringNode).(T), true
	default:
		return zero, false
	}
}
