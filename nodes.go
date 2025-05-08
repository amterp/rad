package rts

import (
	"strconv"

	"github.com/amterp/rts/rsl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

type Shebang struct {
	BaseNode
}

func newShebang(src string, node *ts.Node) (*Shebang, bool) {
	return &Shebang{
		BaseNode: newBaseNode(src, node),
	}, true
}

type FileHeader struct {
	BaseNode
	Contents        string
	MetadataEntries map[string]interface{} // possible values: int, float, string, bool
}

func newFileHeader(src string, node *ts.Node) (*FileHeader, bool) {
	contentsNode := node.ChildByFieldName(rsl.F_CONTENTS)
	if contentsNode == nil {
		// would be strange
		return nil, false
	}

	metadataEntryNodes := node.ChildrenByFieldName(rsl.F_METADATA_ENTRY, node.Walk())

	metadataEntries := make(map[string]interface{}, len(metadataEntryNodes))
	for _, entryNode := range metadataEntryNodes {
		keyNode := entryNode.ChildByFieldName(rsl.F_KEY)
		key := src[keyNode.StartByte():keyNode.EndByte()]
		valueNode := entryNode.ChildByFieldName(rsl.F_VALUE)
		switch valueNode.Kind() {
		case rsl.K_INT:
			str := src[valueNode.StartByte():valueNode.EndByte()]
			value, _ := strconv.Atoi(str)
			metadataEntries[key] = value
		case rsl.K_FLOAT:
			str := src[valueNode.StartByte():valueNode.EndByte()]
			value, _ := strconv.ParseFloat(str, 64)
			metadataEntries[key] = value
		case rsl.K_STRING:
			strContentsNode := valueNode.ChildByFieldName(rsl.F_CONTENTS)
			metadataEntries[key] = src[strContentsNode.StartByte():strContentsNode.EndByte()]
		case rsl.K_BOOL:
			str := src[valueNode.StartByte():valueNode.EndByte()]
			if str == "true" {
				metadataEntries[key] = true
			} else if str == "false" {
				metadataEntries[key] = false
			}
		}
	}

	return &FileHeader{
		BaseNode:        newBaseNode(src, node),
		Contents:        src[contentsNode.StartByte():contentsNode.EndByte()],
		MetadataEntries: metadataEntries,
	}, true
}

type StringNode struct {
	BaseNode
	RawLexeme string // Literal src, excluding delimiters, ws, comments, etc
}

func newStringNode(src string, node *ts.Node) (*StringNode, bool) {
	start := node.ChildByFieldName(rsl.F_START)
	end := node.ChildByFieldName(rsl.F_END)
	contentStart := start.EndByte()
	contentEnd := end.StartByte()
	return &StringNode{
		BaseNode:  newBaseNode(src, node),
		RawLexeme: src[contentStart:contentEnd],
	}, true
}

type CallNode struct {
	BaseNode
	Name     string
	NameNode *ts.Node
}

func newCallNode(node *ts.Node, completeSrc string) (*CallNode, bool) {
	nameNode := node.ChildByFieldName(rsl.F_FUNC)
	if nameNode == nil {
		return nil, false
	}

	name := completeSrc[nameNode.StartByte():nameNode.EndByte()]
	return &CallNode{
		BaseNode: newBaseNode(completeSrc, node),
		Name:     name,
		NameNode: nameNode,
	}, true
}
