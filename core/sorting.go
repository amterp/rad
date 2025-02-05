package core

import "sort"

func sortColumns(
	interp *MainInterpreter,
	token Token,
	fields []Token,
	sorting []ColumnSort,
) {
	orderedCols := make([][]interface{}, 0, len(fields))
	colsByName := make(map[string][]interface{})
	for _, field := range fields {
		values := interp.env.GetByToken(field, RslListT).([]interface{})
		orderedCols = append(orderedCols, values)
		colsByName[field.GetLexeme()] = values
	}

	if len(colsByName) == 0 || len(sorting) == 0 {
		return
	}

	length := len(orderedCols[0])
	if length == 0 {
		return
	}

	indices := make([]int, length)
	for i := range indices {
		indices[i] = i
	}

	sort.Slice(indices, func(i, j int) bool {
		// apply rules in order, breaking ties if needed
		for _, rule := range sorting {
			col := orderedCols[rule.ColIdx]
			comp := compare(interp, token, col[indices[i]], col[indices[j]])
			if comp != 0 {
				return (rule.Dir == Asc && comp < 0) || (rule.Dir == Desc && comp > 0)
			}
		}
		return false
	})

	for name, col := range colsByName {
		sorted := make([]interface{}, length)
		for newIdx, oldIdx := range indices {
			sorted[newIdx] = col[oldIdx]
		}
		interp.env.SetAndImplyTypeWithToken(token, name, sorted)
	}
}

func sortList(interp *MainInterpreter, token Token, data []interface{}, dir SortDir) []interface{} {
	sorted := make([]interface{}, len(data))
	copy(sorted, data)
	sort.Slice(sorted, func(i, j int) bool {
		comp := compare(interp, token, sorted[i], sorted[j])
		if dir == Asc {
			return comp < 0
		}
		return comp > 0
	})
	return sorted
}

func compare(i *MainInterpreter, token Token, a, b interface{}) int {
	// first compare by type
	aTypePrec := precedence(i, token, a)
	bTypePrec := precedence(i, token, b)
	if aTypePrec != bTypePrec {
		if aTypePrec < bTypePrec {
			return -1
		}
		return 1
	}

	// equal type precedence, compare values
	switch aVal := a.(type) {
	case bool:
		bVal := b.(bool)
		if aVal == bVal {
			return 0
		}
		if !aVal && bVal {
			return -1
		}
		return 1
	case int64:
		switch bVal := b.(type) {
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
		switch bVal := b.(type) {
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
		return aVal.Compare(b.(RslString))
	case []interface{}, RslMap:
		return 0 // all arrays and maps are considered equal
	default:
		i.error(token, "Unsupported type for sorting")
		panic(UNREACHABLE)
	}
}

func precedence(i *MainInterpreter, token Token, v interface{}) int {
	switch v.(type) {
	case bool:
		return 0
	case int64, float64:
		return 1
	case RslString:
		return 2
	case []interface{}:
		return 3
	case RslMap:
		return 4
	default:
		i.error(token, "Unsupported type precedence for sorting")
		panic(UNREACHABLE)
	}
}
