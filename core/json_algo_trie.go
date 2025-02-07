package core

import "fmt"

type Trie struct {
	i        *MainInterpreter
	radToken Token
	root     *Node
}

type Node struct {
	key             string
	fullKey         string
	isArrayWildcard bool
	idx             *int64
	// json variables which terminate at this node, and therefore need to capture the data at this level
	fields   []JsonFieldVarOld
	children map[string]*Node
}

type Capture struct {
	node     *Node
	captures map[string][]interface{}
}

func CreateTrie(i *MainInterpreter, radToken Token, jsonFields []JsonFieldVarOld) *Trie {
	trie := &Trie{i: i, radToken: radToken, root: NewNode(nil, "ROOT", false, nil)}
	for _, jsonField := range jsonFields {
		trie.Insert(jsonField)
	}
	return trie
}

func (t *Trie) Insert(field JsonFieldVarOld) {
	node := t.root

	for _, part := range field.Path.Elements {
		key := part.Identifier.GetLexeme()

		if node.children[key] == nil {
			node.children[key] = NewNode(node, key, false, nil)
		}
		node = node.children[key]

		for _, arrElem := range part.ArrElems {
			isArrayWildcard := false
			var index *int64
			if arrElem.ArrayToken != nil {
				isArrayWildcard = true
				key = "[]"
			} else if arrElem.Index != nil {
				idx := (*arrElem.Index).Accept(t.i)
				idxVal, ok := idx.(int64)
				if !ok {
					t.i.error(part.Identifier, fmt.Sprintf("Expected integer index, got %s (%v)", TypeAsString(idx), idx))
				}
				index = &idxVal
				key = fmt.Sprintf("[%d]", idxVal)
			} else {
				t.i.error(part.Identifier, "Bug! Array element has no index or wildcard")
			}

			if node.children[key] == nil {
				node.children[key] = NewNode(node, key, isArrayWildcard, index)
			}
			node = node.children[key]
		}
	}

	node.fields = append(node.fields, field)
}

func NewNode(parent *Node, key string, isArrayWildcard bool, idx *int64) *Node {
	fullKey := ""
	if parent != nil && parent.fullKey != "ROOT" {
		fullKey = parent.fullKey
		if !isArrayWildcard && idx == nil {
			fullKey += "."
		}
	}
	fullKey += key
	return &Node{
		key:             key,
		fullKey:         fullKey,
		isArrayWildcard: isArrayWildcard,
		idx:             idx,
		fields:          []JsonFieldVarOld{},
		children:        make(map[string]*Node),
	}
}
