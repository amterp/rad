package core

import (
	"fmt"
	"strings"

	"github.com/amterp/rad/rts/rl"

	"github.com/samber/lo"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RadMap struct {
	// keys are hashes of a RadValue
	// cannot be collections, though (no maps, lists)
	mapping map[string]RadValue
	keys    []RadValue
}

func NewRadMap() *RadMap {
	return &RadMap{
		mapping: make(map[string]RadValue),
		keys:    []RadValue{},
	}
}

// ShallowCopy creates a shallow copy of the map (keys and values are not deep copied)
func (m *RadMap) ShallowCopy() *RadMap {
	newMap := NewRadMap()
	for _, key := range m.keys {
		value, _ := m.Get(key)
		newMap.Set(key, value)
	}
	return newMap
}

func (m *RadMap) Set(key RadValue, value RadValue) {
	if value == VOID_SENTINEL {
		m.Delete(key)
		return
	}

	if _, exists := m.mapping[key.Hash()]; !exists {
		m.keys = append(m.keys, key)
	}
	m.mapping[key.Hash()] = value
}

func (m *RadMap) SetPrimitiveStr(key string, value string) {
	m.Set(newRadValueStr(key), newRadValueStr(value))
}

func (m *RadMap) SetPrimitiveInt(key string, value int) {
	m.Set(newRadValueStr(key), newRadValueInt(value))
}

func (m *RadMap) SetPrimitiveInt64(key string, value int64) {
	m.Set(newRadValueStr(key), newRadValueInt64(value))
}

func (m *RadMap) SetPrimitiveFloat(key string, value float64) {
	m.Set(newRadValueStr(key), newRadValueFloat64(value))
}

func (m *RadMap) SetPrimitiveBool(key string, value bool) {
	m.Set(newRadValueStr(key), newRadValueBool(value))
}

func (m *RadMap) SetPrimitiveMap(key string, value *RadMap) {
	m.Set(newRadValueStr(key), newRadValueMap(value))
}

func (m *RadMap) SetPrimitiveList(key string, value *RadList) {
	m.Set(newRadValueStr(key), newRadValueList(value))
}

func (m *RadMap) Get(key RadValue) (RadValue, bool) {
	val, exists := m.mapping[key.Hash()]
	return val, exists
}

func (m *RadMap) GetNode(i *Interpreter, idxNode *ts.Node) RadValue {
	// todo grammar: myMap.2 should be okay, treated as "2". but is not valid identifier, so problem!
	if idxNode.Kind() == rl.K_IDENTIFIER {
		// dot syntax e.g. myMap.myKey
		keyName := i.GetSrcForNode(idxNode)
		value, ok := m.Get(newRadValueStr(keyName))
		if !ok {
			// Use panic so fallback operator (??) can catch this error
			errVal := newRadValue(i, idxNode, NewErrorStrf("Key not found: %s", keyName).SetCode(rl.ErrKeyNotFound))
			i.NewRadPanic(idxNode, errVal).Panic()
		}
		return value
	}

	// 'traditional' syntax e.g. myMap["myKey"]
	idxVal := evalMapKey(i, idxNode)
	value, ok := m.Get(idxVal)
	if !ok {
		// todo RAD-138 add mechanism to 'try' getting a key without erroring
		// Use panic so fallback operator (??) can catch this error
		errVal := newRadValue(i, idxNode, NewErrorStrf("Key not found: %s", ToPrintable(idxVal)).SetCode(rl.ErrKeyNotFound))
		i.NewRadPanic(idxNode, errVal).Panic()
	}
	return value
}

func (m *RadMap) Keys() []RadValue {
	return m.keys
}

func (m *RadMap) Values() []RadValue {
	var values []RadValue
	for _, key := range m.keys {
		values = append(values, m.mapping[key.Hash()])
	}
	return values
}

func (m *RadMap) ContainsKey(key RadValue) bool {
	_, exists := m.mapping[key.Hash()]
	return exists
}

func (m *RadMap) Len() int64 {
	return int64(len(m.mapping))
}

func (m *RadMap) Delete(key RadValue) {
	delete(m.mapping, key.Hash())
	// O(n) a little sad but probably okay
	for i, k := range m.keys {
		if k.Equals(key) {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

// fn should return false when it wants to stop. True to continue.
func (m *RadMap) Range(fn func(key, value RadValue) bool) {
	for _, key := range m.keys {
		val := m.mapping[key.Hash()]
		if !fn(key, val) {
			return
		}
	}
}

func (m *RadMap) ToString() string {
	if m.Len() == 0 {
		return "{ }"
	}

	var sb strings.Builder
	sb.WriteString("{ ")

	for i, key := range m.keys {
		value := m.mapping[key.Hash()]
		sb.WriteString(fmt.Sprintf(`%s: %s`, ToPrintable(key), ToPrintable(value)))

		if i < len(m.keys)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(" }")
	return sb.String()
}

func (l *RadMap) Equals(right *RadMap) bool {
	if l.Len() != right.Len() {
		return false
	}

	for _, key := range l.Keys() {
		if val, exists := right.Get(key); !exists {
			return false
		} else if !val.Equals(l.mapping[key.Hash()]) {
			return false
		}
	}

	return true
}

func (m *RadMap) AsErrMsg(i *Interpreter, node *ts.Node) string {
	if lo.Contains(lo.Keys(m.mapping), constCode) && lo.Contains(lo.Keys(m.mapping), constMsg) {
		return fmt.Sprintf("%s (error %s)", m.mapping[constMsg].Val, m.mapping[constCode].Val)
	}

	i.emitErrorf(rl.ErrInternalBug, node, "Bug: Map is not an error message, contains keys: %s", lo.Keys(m.mapping))
	panic(UNREACHABLE)
}

func (m *RadMap) ToGoMap() map[string]interface{} {
	goMap := make(map[string]interface{}, len(m.mapping))
	for k, v := range m.mapping {
		goMap[k] = v.ToGoValue()
	}
	return goMap
}

func evalMapKey(i *Interpreter, idxNode *ts.Node) RadValue {
	return i.eval(idxNode).Val.
		RequireNotType(i, idxNode, "Map keys cannot be lists", rl.RadListT).
		RequireNotType(i, idxNode, "Map keys cannot be maps", rl.RadMapT).
		RequireNotType(i, idxNode, "Map keys cannot be functions", rl.RadFnT)
}
