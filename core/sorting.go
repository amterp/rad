package core

import (
	"sort"

	ts "github.com/tree-sitter/go-tree-sitter"
)

func sortColumns(
	interp *Interpreter,
	fields []radField,
	sorting []ColumnSort,
) {
	orderedCols := make([]*RslList, 0, len(fields))
	colsByName := make(map[radField]*RslList) // todo radField as key works?
	for _, field := range fields {
		colList := interp.env.GetVarElseBug(interp, field.node, field.name).RequireList(interp, field.node)
		orderedCols = append(orderedCols, colList)
		colsByName[field] = colList
	}

	if len(colsByName) == 0 || len(sorting) == 0 {
		return
	}

	length := orderedCols[0].Len()
	if length == 0 {
		return
	}

	// algorithm: we'll sort this 'proxy list' of indices, then apply the resulting
	// sorting to the actual rows
	indices := make([]int64, length)
	for i := range indices {
		indices[i] = int64(i)
	}

	sort.Slice(indices, func(i, j int) bool {
		// apply rules in order, breaking ties if needed
		for _, rule := range sorting {
			colIdx, fieldNode := resolveColIdx(interp, fields, rule.ColIdentifier)
			col := orderedCols[colIdx]
			comp := compare(interp, fieldNode, col.IndexAt(interp, fieldNode, indices[i]), col.IndexAt(interp, fieldNode, indices[j]))
			if comp != 0 {
				return (rule.Dir == Asc && comp < 0) || (rule.Dir == Desc && comp > 0)
			}
		}
		return false
	})

	for field, col := range colsByName {
		col.SortAccordingToIndices(interp, field.node, indices)
	}
}

func resolveColIdx(interp *Interpreter, fields []radField, identifierNode *ts.Node) (int, *ts.Node) {
	identifierStr := interp.sd.Src[identifierNode.StartByte():identifierNode.EndByte()]
	for i, field := range fields {
		if field.name == identifierStr {
			return i, field.node
		}
	}
	interp.errorf(identifierNode, "Bug! Column %q not found", identifierStr)
	panic(UNREACHABLE)
}

func sortList(interp *Interpreter, fieldNode *ts.Node, data *RslList, dir SortDir) []RslValue {
	sorted := make([]RslValue, data.Len())
	copy(sorted, data.Values)
	sort.Slice(sorted, func(i, j int) bool {
		comp := compare(interp, fieldNode, sorted[i], sorted[j])
		if dir == Asc {
			return comp < 0
		}
		return comp > 0
	})
	return sorted
}

func compare(i *Interpreter, fieldNode *ts.Node, a, b RslValue) int {
	// first compare by type
	aTypePrec := precedence(i, fieldNode, a)
	bTypePrec := precedence(i, fieldNode, b)
	if aTypePrec != bTypePrec {
		if aTypePrec < bTypePrec {
			return -1
		}
		return 1
	}

	// equal type precedence, compare values
	switch aVal := a.Val.(type) {
	case bool:
		bVal := b.RequireBool(i, fieldNode)
		if aVal == bVal {
			return 0
		}
		if !aVal && bVal {
			return -1
		}
		return 1
	case int64:
		switch bVal := b.Val.(type) {
		case int64:
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case float64:
			if float64(aVal) < bVal {
				return -1
			}
			if float64(aVal) > bVal {
				return 1
			}
			return 0
		}
		return 0
	case float64:
		switch bVal := b.Val.(type) {
		case float64:
			if aVal < bVal {
				return -1
			}
			if aVal > bVal {
				return 1
			}
			return 0
		case int64:
			if aVal < float64(bVal) {
				return -1
			}
			if aVal > float64(bVal) {
				return 1
			}
			return 0
		}
		return 0
	case RslString:
		return aVal.Compare(b.RequireStr(i, fieldNode))
	case *RslList, *RslMap:
		return 0 // all arrays and maps are considered equal
	default:
		i.errorf(fieldNode, "Bug! Unsupported type for sorting")
		panic(UNREACHABLE)
	}
}

func precedence(i *Interpreter, fieldNode *ts.Node, v RslValue) int {
	switch v.Type() {
	case RslBoolT:
		return 0
	case RslIntT, RslFloatT:
		return 1
	case RslStringT:
		return 2
	case RslListT:
		return 3
	case RslMapT:
		return 4
	default:
		i.errorf(fieldNode, "Unsupported type precedence for sorting")
		panic(UNREACHABLE)
	}
}
