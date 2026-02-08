package core

import (
	"strings"

	"github.com/amterp/rad/rts/rl"
)

type RadList struct {
	Values []RadValue
}

func NewRadList() *RadList {
	return &RadList{
		Values: make([]RadValue, 0),
	}
}

// ShallowCopy creates a shallow copy of the list (values are not deep copied)
func (l *RadList) ShallowCopy() *RadList {
	copiedValues := make([]RadValue, len(l.Values))
	copy(copiedValues, l.Values)
	return &RadList{Values: copiedValues}
}

func NewRadListFromGeneric[T any](i *Interpreter, node rl.Node, list []T) *RadList {
	radList := NewRadList()
	for _, elem := range list {
		radList.Append(newRadValue(i, node, elem))
	}
	return radList
}

func (l *RadList) Append(value RadValue) {
	if value == VOID_SENTINEL {
		return
	}
	l.Values = append(l.Values, value)
}

func (l *RadList) RemoveIdx(i *Interpreter, node rl.Node, idx int) {
	if idx < 0 || idx >= len(l.Values) {
		ErrIndexOutOfBounds(i, node, int64(idx), l.Len())
	}
	l.Values = append(l.Values[:idx], l.Values[idx+1:]...)
}

// IndexAt is intended for internal use
func (l *RadList) IndexAt(i *Interpreter, node rl.Node, idx int64) RadValue {
	if idx < 0 || idx >= l.Len() {
		ErrIndexOutOfBounds(i, node, idx, l.Len())
	}
	return l.Values[idx]
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

func (l *RadList) SortAccordingToIndices(i *Interpreter, node rl.Node, indices []int64) {
	if len(indices) != len(l.Values) {
		i.emitError(rl.ErrInternalBug, node, "Bug: Indices length does not match list length")
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
func (l *RadList) AsActualStringList(i *Interpreter, node rl.Node) []string {
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

func (l *RadList) ToGoList() []interface{} {
	out := make([]interface{}, len(l.Values))
	for idx, val := range l.Values {
		out[idx] = val.ToGoValue()
	}
	return out
}

func (l *RadList) IsEmpty() bool {
	return l.LenInt() == 0
}

func (l *RadList) Reverse() *RadList {
	reversed := NewRadList()
	for i := l.LenInt() - 1; i >= 0; i-- {
		reversed.Append(l.Values[i])
	}
	return reversed
}
