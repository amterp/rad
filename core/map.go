package core

import (
	"fmt"
	"strings"
)

type RslMap struct {
	mapping map[string]interface{}
	keys    []string
}

func NewRslMap() *RslMap {
	return &RslMap{
		mapping: make(map[string]interface{}),
		keys:    []string{},
	}
}

func (m *RslMap) Set(key RslString, value interface{}) {
	m.SetStr(key.Plain(), value)
}

func (m *RslMap) SetStr(key string, value interface{}) {
	if _, exists := m.mapping[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.mapping[key] = value
}

func (m *RslMap) Get(key RslString) (interface{}, bool) {
	return m.GetStr(key.Plain())
}

func (m *RslMap) GetStr(key string) (interface{}, bool) {
	val, exists := m.mapping[key]
	return val, exists
}

func (m *RslMap) Keys() []string {
	return m.keys
}

func (m *RslMap) KeysGeneric() []interface{} {
	var keys []interface{}
	for _, key := range m.keys {
		keys = append(keys, key)
	}
	return keys
}

func (m *RslMap) Values() []interface{} {
	var values []interface{}
	for _, key := range m.keys {
		values = append(values, m.mapping[key])
	}
	return values
}

func (m *RslMap) ContainsKey(key RslString) bool {
	_, exists := m.mapping[key.Plain()]
	return exists
}

func (m *RslMap) Len() int {
	return len(m.mapping)
}

func (m *RslMap) Delete(key RslString) {
	delete(m.mapping, key.Plain())
	// O(n) a little sad but probably okay
	for i, k := range m.keys {
		if k == key.Plain() {
			m.keys = append(m.keys[:i], m.keys[i+1:]...)
			break
		}
	}
}

func (m *RslMap) ToString() string {
	if m.Len() == 0 {
		return "{}"
	}

	var sb strings.Builder
	sb.WriteString("{ ")

	for i, key := range m.keys {
		value := m.mapping[key]
		sb.WriteString(fmt.Sprintf(`%s: %s`, key, ToPrintable(value)))

		if i < len(m.keys)-1 {
			sb.WriteString(", ")
		}
	}

	sb.WriteString(" }")
	return sb.String()
}
