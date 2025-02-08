package core

import (
	"fmt"

	ts "github.com/tree-sitter/go-tree-sitter"
)

type Trie struct {
	i              *Interpreter
	radKeywordNode *ts.Node
	root           *TrieNode
}

type TrieNode struct {
	key             string
	fullKey         string
	isArrayWildcard bool
	idx             *int64
	// json variables which terminate at this node, and therefore need to capture the data at this level
	fields   []JsonFieldVar
	children map[string]*TrieNode
}

type Capture struct {
	node     *TrieNode
	captures map[string][]interface{}
}

func CreateTrie(i *Interpreter, radKeywordNode *ts.Node, jsonFields []JsonFieldVar) *Trie {
	trie := &Trie{
		i:              i,
		radKeywordNode: radKeywordNode,
		root:           NewNode(nil, "ROOT", false, nil),
	}
	for _, jsonField := range jsonFields {
		trie.Insert(jsonField)
	}
	return trie
}

func (t *Trie) Insert(field JsonFieldVar) {
	node := t.root

	for _, segment := range field.Path.Segments {
		key := segment.Identifier

		if node.children[key] == nil {
			node.children[key] = NewNode(node, key, false, nil)
		}
		node = node.children[key]

		for _, idxSegment := range segment.IdxSegments {
			isListWildcard := false
			var index *int64
			if idxSegment.Idx != nil {
				idx := idxSegment.Idx.RequireInt(t.i, idxSegment.IdxNode)
				index = &idx
				key = fmt.Sprintf("[%d]", idx)
			} else if idxSegment.IdxNode != nil {
				isListWildcard = true
				key = "[]"
			}

			if node.children[key] == nil {
				node.children[key] = NewNode(node, key, isListWildcard, index)
			}
			node = node.children[key]
		}
	}

	node.fields = append(node.fields, field)
}

func NewNode(parent *TrieNode, key string, isListWildcard bool, idx *int64) *TrieNode {
	fullKey := ""
	if parent != nil && parent.fullKey != "ROOT" {
		fullKey = parent.fullKey
		if !isListWildcard && idx == nil {
			fullKey += "."
		}
	}
	fullKey += key
	return &TrieNode{
		key:             key,
		fullKey:         fullKey,
		isArrayWildcard: isListWildcard,
		idx:             idx,
		fields:          []JsonFieldVar{},
		children:        make(map[string]*TrieNode),
	}
}
