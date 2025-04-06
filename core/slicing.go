package core

import (
	"github.com/amterp/rts/rsl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

func ResolveSliceStartEnd(i *Interpreter, sliceNode *ts.Node, length int64) (int64, int64) {
	startNode := i.getChild(sliceNode, rsl.F_START)
	endNode := i.getChild(sliceNode, rsl.F_END)

	start := int64(0)
	end := length

	if startNode != nil {
		start = i.evaluate(startNode, 1)[0].RequireInt(i, startNode)
		start = CalculateCorrectedIndex(start, length, true)
	}

	if endNode != nil {
		end = i.evaluate(endNode, 1)[0].RequireInt(i, endNode)
		end = CalculateCorrectedIndex(end, length, true)
	}

	if start > end {
		start = end
	}

	return start, end
}

// 'corrects' negative indices into their positive equivalents
func CalculateCorrectedIndex(rawIdx, length int64, clamp bool) int64 {
	idx := rawIdx
	if rawIdx < 0 {
		idx = rawIdx + length
	}

	if clamp {
		if idx < 0 {
			idx = 0
		} else if idx > length {
			idx = length
		}
	}

	return idx
}
