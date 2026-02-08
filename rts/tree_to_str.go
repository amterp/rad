package rts

import (
	"fmt"
	"strings"

	"github.com/amterp/color"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var escapedReplacer = strings.NewReplacer(
	"\n", "\\n",
	"\r", "\\r",
	"\t", "\\t",
)

func (rt *RadTree) Dump() string {
	root := rt.root.RootNode()
	cursor := root.Walk()
	maxByte, maxPosRow, maxPosCol := findMaxRanges(root, 0, 0, 0, cursor)

	byteLen := len(fmt.Sprintf("%d", maxByte))
	rowLen := len(fmt.Sprintf("%d", maxPosRow))
	colLen := len(fmt.Sprintf("%d", maxPosCol))
	fmtString := fmt.Sprintf("B: [%%%dd, %%%dd] PS: [%%%dd, %%%dd] PE: [%%%dd, %%%dd] %%s%%s%%s",
		byteLen, byteLen, rowLen, colLen, rowLen, colLen)

	var sb strings.Builder
	rt.recurseAppendString(&sb, fmtString, root, "", 0, cursor)

	return sb.String()
}

func findMaxRanges(node *ts.Node, maxByte uint, maxPosRow uint, maxPosCol uint, cursor *ts.TreeCursor) (uint, uint, uint) {
	if node.EndByte() > maxByte {
		maxByte = node.EndByte()
	}
	if node.EndPosition().Row > maxPosRow {
		maxPosRow = node.EndPosition().Row
	}
	if node.EndPosition().Column > maxPosCol {
		maxPosCol = node.EndPosition().Column
	}

	children := node.Children(cursor)
	for _, child := range children {
		maxByte, maxPosRow, maxPosCol = findMaxRanges(&child, maxByte, maxPosRow, maxPosCol, cursor)
	}
	return maxByte, maxPosRow, maxPosCol
}

func (rt *RadTree) recurseAppendString(
	sb *strings.Builder,
	fmtString string,
	node *ts.Node,
	nodeFieldName string,
	treeLevel int,
	cursor *ts.TreeCursor,
) {
	indent := strings.Repeat("  ", treeLevel)
	if nodeFieldName != "" {
		nodeFieldName = color.MagentaString(nodeFieldName)
		nodeFieldName += ": "
	}

	var nodeKind string
	if node.IsError() {
		nodeKind = color.RedString("ERROR")
	} else {
		nodeKind = color.GreenString(node.Kind())
	}

	sb.WriteString(fmt.Sprintf(fmtString,
		node.StartByte(), node.EndByte(),
		node.StartPosition().Row, node.StartPosition().Column,
		node.EndPosition().Row, node.EndPosition().Column,
		indent,
		nodeFieldName,
		nodeKind,
	))

	if node.IsMissing() {
		sb.WriteString(color.RedString(" (MISSING)"))
	}

	children := node.Children(cursor)
	if len(children) == 0 {
		src := rt.src[node.StartByte():node.EndByte()]
		sb.WriteString(fmt.Sprintf(" `%s`\n", color.YellowString(escapedReplacer.Replace(src))))
		return
	} else {
		sb.WriteString("\n")
	}

	for i, child := range children {
		childFieldName := node.FieldNameForChild(uint32(i))
		rt.recurseAppendString(sb, fmtString, &child, childFieldName, treeLevel+1, cursor)
	}
}
