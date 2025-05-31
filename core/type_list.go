package core

import (
	"strings"

	"github.com/amterp/rad/rts/rl"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadList struct {
	Values []RadValue
}

func NewRadList() *RadList {
	return &RadList{
		Values: make([]RadValue, 0),
	}
}

func NewRadListFromGeneric[T any](i *Interpreter, node *ts.Node, list []T) *RadList {
	radList := NewRadList()
	for _, elem := range list {
		radList.Append(newRadValue(i, node, elem))
	}
	return radList
}

func (l *RadList) Append(value RadValue) {
	l.Values = append(l.Values, value)
}

func (l *RadList) GetIdx(i *Interpreter, idxNode *ts.Node) RadValue {
	if idxNode.Kind() == rl.K_SLICE {
		return newRadValue(i, idxNode, l.Slice(i, idxNode))
	}

	idxVal := i.evaluate(idxNode, 1)[0]
	rawIdx := idxVal.RequireInt(i, idxNode)
	idx := CalculateCorrectedIndex(rawIdx, l.Len(), false)
	if idx < 0 || idx >= l.Len() {
		ErrIndexOutOfBounds(i, idxNode, rawIdx, l.Len())
	}
	return l.Values[idx]
}

func (l *RadList) ModifyIdx(i *Interpreter, idxNode *ts.Node, value RadValue) {
	if idxNode.Kind() == rl.K_SLICE {
		start, end := ResolveSliceStartEnd(i, idxNode, l.Len())
		if start < end {
			newList := NewRadList()
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

func (l *RadList) RemoveIdx(i *Interpreter, node *ts.Node, idx int) {
	if idx < 0 || idx >= len(l.Values) {
		ErrIndexOutOfBounds(i, node, int64(idx), l.Len())
	}
	l.Values = append(l.Values[:idx], l.Values[idx+1:]...)
}

// more intended for internal use than GetIdx
func (l *RadList) IndexAt(i *Interpreter, node *ts.Node, idx int64) RadValue {
	if idx < 0 || idx >= l.Len() {
		ErrIndexOutOfBounds(i, node, idx, l.Len())
	}
	return l.Values[idx]
}

func (l *RadList) Slice(i *Interpreter, sliceNode *ts.Node) *RadList {
	start, end := ResolveSliceStartEnd(i, sliceNode, l.Len())

	newList := NewRadList()
	for i := start; i < end; i++ {
		newList.Append(l.Values[i])
	}

	return newList
}

func (l *RadList) Contains(val RadValue) bool {
	for _, elem := range l.Values {
		if elem.Equals(val) {
			return true
		}
	}
	return false
}

func (l *RadList) JoinWith(other *RadList) RadValue {
	newList := NewRadList()
	newList.Values = append(l.Values, other.Values...)
	return newRadValue(nil, nil, newList)
}

func (l *RadList) Equals(r *RadList) bool {
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

func (l *RadList) ToString() string {
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

func (l *RadList) Len() int64 {
	return int64(l.LenInt())
}

func (l *RadList) LenInt() int {
	return len(l.Values)
}

func (l *RadList) SortAccordingToIndices(i *Interpreter, node *ts.Node, indices []int64) {
	if len(indices) != len(l.Values) {
		i.errorf(node, "Bug! Indices length does not match list length")
	}
	sorted := make([]RadValue, l.Len())
	for newIdx, oldIdx := range indices {
		sorted[newIdx] = l.Values[oldIdx]
	}
	l.Values = sorted
}

func (l *RadList) AsStringList(quoteStrings bool) []string {
	out := make([]string, l.Len())
	for i, elem := range l.Values {
		out[i] = ToPrintableQuoteStr(elem, quoteStrings)
	}
	return out
}

// requires contents to actually be strings
func (l *RadList) AsActualStringList(i *Interpreter, node *ts.Node) []string {
	out := make([]string, l.Len())
	for idx, elem := range l.Values {
		out[idx] = elem.RequireStr(i, node).Plain() // todo keep attributes?
	}
	return out
}

func (l *RadList) Join(sep string, prefix string, suffix string) RadString {
	var arr []string
	for _, v := range l.Values {
		arr = append(arr, ToPrintableQuoteStr(v, false))
	}
	return NewRadString(prefix + strings.Join(arr, sep) + suffix)
}
