package rts

import (
	"errors"
	"fmt"

	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadTree struct {
	root *ts.Tree
	// Updatable:
	parser *ts.Parser
	src    string
}

func newRadTree(parser *ts.Parser, tree *ts.Tree, src string) *RadTree {
	return &RadTree{
		root:   tree,
		parser: parser,
		src:    src,
	}
}

func (rt *RadTree) Update(src string) {
	// todo use incrΩemental parsing, maybe can lean on LSP client to give via protocol
	rt.root = rt.parser.Parse([]byte(src), nil)
	rt.src = src
}

func (rt *RadTree) Close() {
	rt.root.Close()
}

func (rt *RadTree) Root() *ts.Node {
	return rt.root.RootNode()
}

func (rt *RadTree) Sexp() string {
	return rt.root.RootNode().ToSexp()
}

func (rt *RadTree) String() string {
	return rt.Dump()
}

func (rt *RadTree) FindShebang() (*Shebang, bool) {
	shebangs := findNodes[*Shebang](rt)
	if len(shebangs) == 0 {
		return nil, false
	}
	return shebangs[0], true // todo bad if multiple
}

func (rt *RadTree) FindFileHeader() (*FileHeader, bool) {
	fileHeaders := findNodes[*FileHeader](rt)
	if len(fileHeaders) == 0 {
		return nil, false
	}
	return fileHeaders[0], true // todo bad if multiple
}

func (rt *RadTree) FindArgBlock() (*ArgBlock, bool) {
	argBlocks := findNodes[*ArgBlock](rt)
	if len(argBlocks) == 0 {
		return nil, false
	}
	return argBlocks[0], true // todo bad if multiple
}

func (rt *RadTree) FindCmdBlocks() ([]*CmdBlock, bool) {
	// Commands can appear multiple times, so we need to find all of them
	nodes := rt.FindNodes(rl.K_CMD_BLOCK)
	if len(nodes) == 0 {
		return nil, false
	}

	cmdBlocks := make([]*CmdBlock, 0, len(nodes))
	for _, node := range nodes {
		cmdBlock, ok := newCmdBlock(rt.src, node)
		if ok {
			cmdBlocks = append(cmdBlocks, cmdBlock)
		}
	}

	return cmdBlocks, len(cmdBlocks) > 0
}

func QueryNodes[T Node](rt *RadTree) ([]T, error) {
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

func (rt *RadTree) FindInvalidNodes() []*ts.Node {
	var invalidNodes []*ts.Node
	root := rt.Root()
	cursor := root.Walk()
	recurseFindInvalidNodes(root, &invalidNodes, cursor)
	return invalidNodes
}

func (rt *RadTree) FindCalls() []*CallNode {
	calls := rt.FindNodes(rl.K_CALL)
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
func (rt *RadTree) FindNodes(nodeKind string) []*ts.Node {
	var found []*ts.Node
	root := rt.Root()
	cursor := root.Walk()
	recurseFindNodes(root, nodeKind, &found, cursor)
	return found
}

func recurseFindNodes(node *ts.Node, nodeKind string, found *[]*ts.Node, cursor *ts.TreeCursor) {
	if node.Kind() == nodeKind {
		*found = append(*found, node)
	}
	childrenNodes := node.Children(cursor)
	for _, child := range childrenNodes {
		recurseFindNodes(&child, nodeKind, found, cursor)
	}
}

func recurseFindInvalidNodes(node *ts.Node, invalidNodes *[]*ts.Node, cursor *ts.TreeCursor) {
	if node.IsError() || node.IsMissing() {
		*invalidNodes = append(*invalidNodes, node)
	}
	childrenNodes := node.Children(cursor)
	for _, child := range childrenNodes {
		recurseFindInvalidNodes(&child, invalidNodes, cursor)
	}
}

func findNodes[T Node](rt *RadTree) []T {
	nodeName := NodeName[T]()
	rootNode := rt.root.RootNode()
	cursor := rootNode.Walk()
	node, ok := rt.findFirstNode(nodeName, rootNode, cursor)
	if !ok {
		return []T{}
	}
	rtsNode, ok := createNode[T](rt.src, node)
	if !ok {
		panic(errors.New("failed to create RTS version of node"))
	}
	return []T{rtsNode} // todo stub - should search all
}

func (rt *RadTree) findFirstNode(nodeKind string, node *ts.Node, cursor *ts.TreeCursor) (*ts.Node, bool) {
	if node.Kind() == nodeKind {
		return node, true
	}
	children := node.Children(cursor)
	for _, child := range children {
		if n, ok := rt.findFirstNode(nodeKind, &child, cursor); ok {
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
	case *CmdBlock:
		cmdBlock, _ := newCmdBlock(src, node)
		return any(cmdBlock).(T), true
	case *StringNode:
		stringNode, _ := newStringNode(src, node)
		return any(stringNode).(T), true
	default:
		return zero, false
	}
}
