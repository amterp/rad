package core

import (
	"github.com/amterp/rad/rts/rl"
	ts "github.com/tree-sitter/go-tree-sitter"
)

func ResolveSliceStartEnd(i *Interpreter, sliceNode *ts.Node, length int64) (int64, int64) {
	startNode := rl.GetChild(sliceNode, rl.F_START)
	endNode := rl.GetChild(sliceNode, rl.F_END)

	start := int64(0)
	end := length

	if startNode != nil {
		start = i.eval(startNode).Val.RequireInt(i, startNode)
		start = CalculateCorrectedIndex(start, length, true)
	}

	if endNode != nil {
		end = i.eval(endNode).Val.RequireInt(i, endNode)
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
