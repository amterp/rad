package core

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

func ResolveSliceStartEnd(i *Interpreter, startNode *ts.Node, endNode *ts.Node, length int64) (int64, int64) {
	start := int64(0)
	end := length

	if startNode != nil {
		start = i.evaluate(startNode, 1)[0].RequireInt(i, startNode)
		start = resolveSliceIndex(start, length)
	}

	if endNode != nil {
		end = i.evaluate(endNode, 1)[0].RequireInt(i, endNode)
		end = resolveSliceIndex(end, length)
	}

	return start, end
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
