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

func (l *RslList) Append(value RslValue) {
	l.Values = append(l.Values, value)
}

// todo support negative indices
func (l *RslList) GetIdx(i *Interpreter, idxNode *ts.Node) RslValue {
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
