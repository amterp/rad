package core

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

func (l *RslList) Slice(i *Interpreter, startNode, endNode *ts.Node) *RslList {
	start := int64(0)
	listLen := l.Len()
	end := listLen

	if startNode != nil {
		start = i.evaluate(startNode, 1)[0].RequireInt(i, startNode)
		start = resolveSliceIndex(start, listLen)
	}

	if endNode != nil {
		end = i.evaluate(endNode, 1)[0].RequireInt(i, endNode)
		end = resolveSliceIndex(end, listLen)
	}

	newList := NewRslList()
	for i := start; i < end; i++ {
		newList.Append(l.Values[i])
	}

	return newList
}

func resolveSliceIndex(rawIdx, listLen int64) int64 {
	idx := rawIdx
	if rawIdx < 0 {
		idx = rawIdx + listLen
	}

	if idx < 0 {
		idx = 0
	} else if idx > listLen {
		idx = listLen
	}

	return idx
}
