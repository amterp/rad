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

func (m *RslMap) Set(key string, value interface{}) {
	if _, exists := m.mapping[key]; !exists {
		m.keys = append(m.keys, key)
	}
	m.mapping[key] = value
}

func (m *RslMap) Get(key string) (interface{}, bool) {
	val, exists := m.mapping[key]
	return val, exists
}

func (m *RslMap) Keys() []string {
	return m.keys
}

func (m *RslMap) Len() int {
	return len(m.mapping)
}

func (m *RslMap) Delete(key string) {
	delete(m.mapping, key)
	// O(n) a little sad but probably okay
	for i, k := range m.keys {
		if k == key {
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
