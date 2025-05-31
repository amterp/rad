package core

import (
	"strings"

	"github.com/amterp/rad/rts/rsl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslList struct {
	Values []RslValue
}

func NewRslList() *RslList {
	return &RslList{
		Values: make([]RslValue, 0),
	}
}

func NewRslListFromGeneric[T any](i *Interpreter, node *ts.Node, list []T) *RslList {
	rslList := NewRslList()
	for _, elem := range list {
		rslList.Append(newRslValue(i, node, elem))
	}
	return rslList
}

func (l *RslList) Append(value RslValue) {
	l.Values = append(l.Values, value)
}

func (l *RslList) GetIdx(i *Interpreter, idxNode *ts.Node) RslValue {
	if idxNode.Kind() == rsl.K_SLICE {
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
	if idxNode.Kind() == rsl.K_SLICE {
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

func (l *RslList) RemoveIdx(i *Interpreter, node *ts.Node, idx int) {
	if idx < 0 || idx >= len(l.Values) {
		ErrIndexOutOfBounds(i, node, int64(idx), l.Len())
	}
	l.Values = append(l.Values[:idx], l.Values[idx+1:]...)
}

// more intended for internal use than GetIdx
func (l *RslList) IndexAt(i *Interpreter, node *ts.Node, idx int64) RslValue {
	if idx < 0 || idx >= l.Len() {
		ErrIndexOutOfBounds(i, node, idx, l.Len())
	}
	return l.Values[idx]
}

func (l *RslList) Slice(i *Interpreter, sliceNode *ts.Node) *RslList {
	start, end := ResolveSliceStartEnd(i, sliceNode, l.Len())

	newList := NewRslList()
	for i := start; i < end; i++ {
		newList.Append(l.Values[i])
	}

	return newList
}

func (l *RslList) Contains(val RslValue) bool {
	for _, elem := range l.Values {
		if elem.Equals(val) {
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
	return int64(l.LenInt())
}

func (l *RslList) LenInt() int {
	return len(l.Values)
}

func (l *RslList) SortAccordingToIndices(i *Interpreter, node *ts.Node, indices []int64) {
	if len(indices) != len(l.Values) {
		i.errorf(node, "Bug! Indices length does not match list length")
	}
	sorted := make([]RslValue, l.Len())
	for newIdx, oldIdx := range indices {
		sorted[newIdx] = l.Values[oldIdx]
	}
	l.Values = sorted
}

func (l *RslList) AsStringList(quoteStrings bool) []string {
	out := make([]string, l.Len())
	for i, elem := range l.Values {
		out[i] = ToPrintableQuoteStr(elem, quoteStrings)
	}
	return out
}

// requires contents to actually be strings
func (l *RslList) AsActualStringList(i *Interpreter, node *ts.Node) []string {
	out := make([]string, l.Len())
	for idx, elem := range l.Values {
		out[idx] = elem.RequireStr(i, node).Plain() // todo keep attributes?
	}
	return out
}

func (l *RslList) Join(sep string, prefix string, suffix string) RslString {
	var arr []string
	for _, v := range l.Values {
		arr = append(arr, ToPrintableQuoteStr(v, false))
	}
	return NewRslString(prefix + strings.Join(arr, sep) + suffix)
}
