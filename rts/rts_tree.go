package rts

import (
	"errors"
	"fmt"
	"sync"

	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

// RadTree wraps a tree-sitter *ts.Tree, which owns C-heap memory via
// cgo. The C memory must be released by calling ts.Tree.Close(); the
// Go garbage collector alone won't free it.
//
// We do NOT register a finalizer on RadTree. The upstream
// go-tree-sitter README warns against it, and we hit the documented
// hazard during development: tree-sitter Nodes hold references into
// the C tree but the Go-side relationship between Node and Tree isn't
// visible to the GC. A finalizer-driven Close could free the C tree
// while another goroutine is walking nodes that still point into it
// (the converter does exactly this).
//
// Cleanup is the caller's responsibility: call Close when no further
// readers will touch the tree. Higher layers (e.g. radls's snapshot
// model) coordinate that lifetime via reference counting.
//
// closeOnce makes Close idempotent so callers can defer-Close without
// worrying about double-free; upstream ts.Tree.Close is NOT
// idempotent (it always calls ts_tree_delete), so we guard.
//
// We hold only the language (immutable, safe to share across
// goroutines), not the parser. Any reparse path threads the parser
// in as a method argument so the same parser can be locked-and-
// driven by the owning State without RadTree carrying a hazardous
// back-pointer.
type RadTree struct {
	root     *ts.Tree
	language *ts.Language
	src      string

	closeOnce sync.Once
}

func newRadTree(language *ts.Language, tree *ts.Tree, src string) *RadTree {
	return &RadTree{
		root:     tree,
		language: language,
		src:      src,
	}
}

// Update reparses the tree in place with new source, using the
// supplied parser. Used by callers that hold a long-lived RadTree
// (e.g. the check package's standalone driver). The snapshot model
// in radls doesn't use this - each version gets a fresh tree - so
// this path is reserved for non-snapshot consumers. Closes the
// previous tree to avoid leaking C memory.
//
// Taking the parser as an argument (rather than a struct field) means
// the caller controls parser lifetime and concurrency - a RadTree
// can't accidentally race a shared parser through its own back-pointer.
func (rt *RadTree) Update(parser *RadParser, src string) {
	// todo use incremental parsing, maybe can lean on LSP client to give via protocol
	old := rt.root
	rt.root = parser.parser.Parse([]byte(src), nil)
	rt.src = src
	if old != nil {
		old.Close()
	}
}

// Close releases the underlying tree-sitter C memory. Idempotent;
// the underlying ts.Tree.Close is not, so we guard with sync.Once.
// Safe to call from multiple goroutines.
func (rt *RadTree) Close() {
	rt.closeOnce.Do(func() {
		if rt.root != nil {
			rt.root.Close()
			rt.root = nil
		}
	})
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

func QueryNodes[T Node](rt *RadTree) ([]T, error) {
	nodeName := NodeName[T]()
	query, err := ts.NewQuery(rt.language, fmt.Sprintf("(%s) @%s", nodeName, nodeName))

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
	recurseFindInvalidNodes(rt.Root(), &invalidNodes)
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

// SpanFromNode creates an rl.Span from a tree-sitter node and file path.
func SpanFromNode(node *ts.Node, file string) rl.Span {
	return rl.Span{
		File:      file,
		StartByte: int(node.StartByte()),
		EndByte:   int(node.EndByte()),
		StartRow:  int(node.StartPosition().Row),
		StartCol:  int(node.StartPosition().Column),
		EndRow:    int(node.EndPosition().Row),
		EndCol:    int(node.EndPosition().Column),
	}
}

// FindInvalidNodeSpans returns spans for all invalid/missing nodes.
func (rt *RadTree) FindInvalidNodeSpans(file string) []rl.Span {
	nodes := rt.FindInvalidNodes()
	spans := make([]rl.Span, len(nodes))
	for i, n := range nodes {
		spans[i] = SpanFromNode(n, file)
	}
	return spans
}

// HasInvalidNodes returns true if the tree contains any error/missing nodes.
func (rt *RadTree) HasInvalidNodes() bool {
	return len(rt.FindInvalidNodes()) > 0
}

func findNodes[T Node](rt *RadTree) []T {
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

func (rt *RadTree) findFirstNode(nodeKind string, node *ts.Node) (*ts.Node, bool) {
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
	case *StringNode:
		stringNode, _ := newStringNode(src, node)
		return any(stringNode).(T), true
	default:
		return zero, false
	}
}
