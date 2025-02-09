package core

import (
	"fmt"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type RslMap struct {
	// keys are hashes of an RslValue
	// cannot be collections, though (no maps, lists)
	mapping map[string]RslValue
	keys    []RslValue
}

func NewRslMap() *RslMap {
	return &RslMap{
		mapping: make(map[string]RslValue),
		keys:    []RslValue{},
	}
}

func (m *RslMap) Set(key RslValue, value RslValue) {
	if value == NIL_SENTINAL {
		m.Delete(key)
		return
	}

	if _, exists := m.mapping[key.Hash()]; !exists {
		m.keys = append(m.keys, key)
	}
	m.mapping[key.Hash()] = value
}

func (m *RslMap) SetPrimitiveStr(key string, value string) {
	m.Set(newRslValueStr(key), newRslValueStr(value))
}

func (m *RslMap) SetPrimitiveInt(key string, value int) {
	m.Set(newRslValueStr(key), newRslValueInt(value))
}

func (m *RslMap) SetPrimitiveInt64(key string, value int64) {
	m.Set(newRslValueStr(key), newRslValueInt64(value))
}

func (m *RslMap) SetPrimitiveMap(key string, value *RslMap) {
	m.Set(newRslValueStr(key), newRslValueMap(value))
}

func (m *RslMap) Get(key RslValue) (RslValue, bool) {
	val, exists := m.mapping[key.Hash()]
	return val, exists
}

func (m *RslMap) GetNode(i *Interpreter, idxNode *ts.Node) RslValue {
	// todo grammar: myMap.2 should be okay, treated as "2". but is not valid identifier, so problem!
	if idxNode.Kind() == K_IDENTIFIER {
		// dot syntax e.g. myMap.myKey
		keyName := i.sd.Src[idxNode.StartByte():idxNode.EndByte()]
		value, ok := m.Get(newRslValueStr(keyName))
		if !ok {
			i.errorf(idxNode, "Key not found: %s", keyName)
		}
		return value
	}

	// 'traditional' syntax e.g. myMap["myKey"]
	idxVal := evalMapKey(i, idxNode)
	value, ok := m.Get(idxVal)
	if !ok {
		// todo RAD-138 add mechanism to 'try' getting a key without erroring
		i.errorf(idxNode, "Key not found: %s", idxVal)
	}
	return value
}

func (m *RslMap) Keys() []RslValue {
	return m.keys
}

func (m *RslMap) Values() []RslValue {
	var values []RslValue
	for _, key := range m.keys {
		values = append(values, m.mapping[key.Hash()])
	}
	return values
}

func (m *RslMap) ContainsKey(key RslValue) bool {
	_, exists := m.mapping[key.Hash()]
	return exists
}

func (m *RslMap) Len() int64 {
	return int64(len(m.mapping))
}

func (m *RslMap) Delete(key RslValue) {
	delete(m.mapping, key.Hash())
	// O(n) a little sad but probably okay
	for i, k := range m.keys {
		if k.Equals(key) {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *RslMap) ToString() string {
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

func (l *RslMap) Equals(right *RslMap) bool {
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

func evalMapKey(i *Interpreter, idxNode *ts.Node) RslValue {
	return i.evaluate(idxNode, 1)[0].
		RequireNotType(i, idxNode, "Map keys cannot be lists", RslListT).
		RequireNotType(i, idxNode, "Map keys cannot be maps", RslMapT)
}
