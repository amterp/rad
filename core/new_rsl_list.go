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

// todo support negative indices
func (l *RslList) GetIdx(i *Interpreter, idxNode *ts.Node) RslValue {
	if idxNode.Kind() == K_SLICE {
		return newRslValue(i, idxNode, l.Slice(i, i.getChild(idxNode, F_START), i.getChild(idxNode, F_END)))
	}

	idxVal := i.evaluate(idxNode, 1)[0]
	idxInt := idxVal.RequireInt(i, idxNode)
	if idxInt < 0 || idxInt >= int64(len(l.Values)) {
		i.errorf(idxNode, "Index out of bounds: %d", idxInt)
	}
	return l.Values[idxInt]
}

// todo support negative indices
func (l *RslList) ModifyIdx(i *Interpreter, idxNode *ts.Node, value RslValue) {
	idxVal := i.evaluate(idxNode, 1)[0]
	idxInt := idxVal.RequireInt(i, idxNode)
	if idxInt < 0 || idxInt >= int64(len(l.Values)) {
		i.errorf(idxNode, "Index out of bounds: %d", idxInt)
	}
	l.Values[idxInt] = value
}

func (l *RslList) Slice(i *Interpreter, startNode, endNode *ts.Node) *RslList {
	start, end := ResolveSliceStartEnd(i, startNode, endNode, l.Len())

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
