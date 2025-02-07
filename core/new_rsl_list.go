package core

import ts "github.com/tree-sitter/go-tree-sitter"

type RslList struct {
	Values []RslValue
}

func NewRslList() *RslList {
	return &RslList{
		Values: make([]RslValue, 0),
	}
}

func NewFromGeneric[T any](list []T) *RslList {
	rslList := NewRslList()
	for _, elem := range list {
		rslList.Append(newRslValue(nil, nil, elem))
	}
	return rslList
}

func (l *RslList) Append(value RslValue) {
	l.Values = append(l.Values, value)
}

func (l *RslList) GetIdx(i *Interpreter, idxNode *ts.Node) RslValue {
	if idxNode.Kind() == K_SLICE {
		return newRslValue(i, idxNode, l.Slice(i, idxNode))
	}

	idxVal := i.evaluate(idxNode, 1)[0]
	rawIdx := idxVal.RequireInt(i, idxNode)
	idx := CalculateCorrectedIndex(rawIdx, l.Len(), false)
	if idx < 0 || idx >= l.Len() {
		ErrIndexOutOfBounds(i, idxNode, rawIdx, l.Len())
	}
	return l.Values[idx]
}

func (l *RslList) ModifyIdx(i *Interpreter, idxNode *ts.Node, value RslValue) {
	if idxNode.Kind() == K_SLICE {
		start, end := ResolveSliceStartEnd(i, idxNode, l.Len())
		if start < end {
			newList := NewRslList()
			newList.Values = append(newList.Values, l.Values[:start]...)
			if list, ok := value.TryGetList(); ok {
				newList.Values = append(newList.Values, list.Values...)
			} else if value == NIL_SENTINAL {
				// do nothing (delete those values)
			} else {
				assignNode := idxNode.Parent().Parent()
				i.errorf(assignNode, "Cannot assign list slice to a non-list type")
			}
			newList.Values = append(newList.Values, l.Values[end:]...)
			l.Values = newList.Values
		}
		return
	}

	// regular single index

	idxVal := i.evaluate(idxNode, 1)[0]
	rawIdx := idxVal.RequireInt(i, idxNode)
	idx := CalculateCorrectedIndex(rawIdx, l.Len(), false)
	if idx < 0 || idx >= int64(len(l.Values)) {
		ErrIndexOutOfBounds(i, idxNode, rawIdx, l.Len())
	}

	if value == NIL_SENTINAL {
		l.Values = append(l.Values[:idx], l.Values[idx+1:]...)
	} else {
		l.Values[idx] = value
	}
}

func (l *RslList) Slice(i *Interpreter, sliceNode *ts.Node) *RslList {
	start, end := ResolveSliceStartEnd(i, sliceNode, l.Len())

	newList := NewRslList()
	for i := start; i < end; i++ {
		newList.Append(l.Values[i])
	}

	return newList
}

func (l *RslList) Contains(val interface{}) bool {
	for _, elem := range l.Values {
		if elem.Val == val {
			return true
		}
	}
	return false
}

func (l *RslList) JoinWith(other *RslList) RslValue {
	newList := NewRslList()
	newList.Values = append(l.Values, other.Values...)
	return newRslValue(nil, nil, newList)
}

func (l *RslList) Equals(r *RslList) bool {
	if len(l.Values) != len(r.Values) {
		return false
	}
	for i, elem := range l.Values {
		if !elem.Equals(r.Values[i]) {
			return false
		}
	}
	return true
}

func (l *RslList) ToString() string {
	if l.Len() == 0 {
		return "[ ]"
	}

	out := "[ "
	for i, elem := range l.Values {
		if i > 0 {
			out += ", "
		}
		out += ToPrintable(elem)
	}
	return out + " ]"
}

func (l *RslList) Len() int64 {
	return int64(len(l.Values))
}
