package rts

import (
	"errors"
	"fmt"

	"github.com/amterp/rad/rts/rsl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslTree struct {
	root *ts.Tree
	// Updatable:
	parser *ts.Parser
	src    string
}

func newRslTree(parser *ts.Parser, tree *ts.Tree, src string) *RslTree {
	return &RslTree{
		root:   tree,
		parser: parser,
		src:    src,
	}
}

func (rt *RslTree) Update(src string) {
	// todo use incremental parsing, maybe can lean on LSP client to give via protocol
	rt.root = rt.parser.Parse([]byte(src), nil)
	rt.src = src
}

func (rt *RslTree) Close() {
	rt.root.Close()
}

func (rt *RslTree) Root() *ts.Node {
	return rt.root.RootNode()
}

func (rt *RslTree) Sexp() string {
	return rt.root.RootNode().ToSexp()
}

func (rt *RslTree) String() string {
	return rt.Dump()
}

func (rt *RslTree) FindShebang() (*Shebang, bool) {
	shebangs := findNodes[*Shebang](rt)
	if len(shebangs) == 0 {
		return nil, false
	}
	return shebangs[0], true // todo bad if multiple
}

func (rt *RslTree) FindFileHeader() (*FileHeader, bool) {
	fileHeaders := findNodes[*FileHeader](rt)
	if len(fileHeaders) == 0 {
		return nil, false
	}
	return fileHeaders[0], true // todo bad if multiple
}

func (rt *RslTree) FindArgBlock() (*ArgBlock, bool) {
	argBlocks := findNodes[*ArgBlock](rt)
	if len(argBlocks) == 0 {
		return nil, false
	}
	return argBlocks[0], true // todo bad if multiple
}

func QueryNodes[T Node](rt *RslTree) ([]T, error) {
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

func (rt *RslTree) FindInvalidNodes() []*ts.Node {
	var invalidNodes []*ts.Node
	recurseFindInvalidNodes(rt.Root(), &invalidNodes)
	return invalidNodes
}

func (rt *RslTree) FindCalls() []*CallNode {
	calls := rt.FindNodes(rsl.K_CALL)
	callNodes := make([]*CallNode, len(calls))
	for i, call := range calls {
		callNode, ok := newCallNode(call, rt.src)
		if !ok {
			panic(errors.New("failed to create RTS version of node"))
		}
		callNodes[i] = callNode
	}
	return callNodes
}

// todo should take an ID instead of string for kind
func (rt *RslTree) FindNodes(nodeKind string) []*ts.Node {
	var found []*ts.Node
	recurseFindNodes(rt.Root(), nodeKind, &found)
	return found
}

func recurseFindNodes(node *ts.Node, nodeKind string, found *[]*ts.Node) {
	if node.Kind() == nodeKind {
		*found = append(*found, node)
	}
	childrenNodes := node.Children(node.Walk())
	for _, child := range childrenNodes {
		recurseFindNodes(&child, nodeKind, found)
	}
}

func recurseFindInvalidNodes(node *ts.Node, invalidNodes *[]*ts.Node) {
	if node.IsError() || node.IsMissing() {
		*invalidNodes = append(*invalidNodes, node)
	}
	childrenNodes := node.Children(node.Walk())
	for _, child := range childrenNodes {
		recurseFindInvalidNodes(&child, invalidNodes)
	}
}

func findNodes[T Node](rt *RslTree) []T {
	nodeName := NodeName[T]()
	node, ok := rt.findFirstNode(nodeName, rt.root.RootNode())
	if !ok {
		return []T{}
	}
	rtsNode, ok := createNode[T](rt.src, node)
	if !ok {
		panic(errors.New("failed to create RTS version of node"))
	}
	return []T{rtsNode} // todo stub - should search all
}

func (rt *RslTree) findFirstNode(nodeKind string, node *ts.Node) (*ts.Node, bool) {
	if node.Kind() == nodeKind {
		return node, true
	}
	children := node.Children(node.Walk())
	for _, child := range children {
		if n, ok := rt.findFirstNode(nodeKind, &child); ok {
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
	case *ArgBlock:
		argBlock, _ := newArgBlock(src, node)
		return any(argBlock).(T), true
	case *StringNode:
		stringNode, _ := newStringNode(src, node)
		return any(stringNode).(T), true
	default:
		return zero, false
	}
}
