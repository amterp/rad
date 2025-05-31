package rts

import (
	"regexp"
	"strings"

	"github.com/amterp/rad/rts/rsl"
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
	MetadataEntries map[string]string
}

func newFileHeader(src string, node *ts.Node) (*FileHeader, bool) {
	contentsNode := node.ChildByFieldName(rsl.F_CONTENTS)
	if contentsNode == nil {
		// would be strange
		return nil, false
	}

	rawContents := src[contentsNode.StartByte():contentsNode.EndByte()]
	fhContents, metadataEntries := extractContentsAndMetadata(rawContents)

	return &FileHeader{
		BaseNode:        newBaseNode(src, node),
		Contents:        fhContents,
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

func extractContentsAndMetadata(rawContents string) (string, map[string]string) {
	lines := strings.Split(rawContents, "\n")

	metadataRegex := regexp.MustCompile(`^@([a-zA-Z0-9_]+)\s*=\s*(.+)$`)

	metadataEntries := make(map[string]string)
	var fhContents string
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])

		if line == "" {
			continue
		}

		matches := metadataRegex.FindStringSubmatch(line)

		if matches != nil && len(matches) == 3 {
			identifier := matches[1]
			value := strings.TrimSpace(matches[2])
			metadataEntries[identifier] = value
		} else {
			lines = lines[:i+1]
			fhContents = strings.Join(lines, "\n")
			break
		}
	}
	return fhContents, metadataEntries
}
