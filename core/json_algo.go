package core

import (
	"fmt"
	"github.com/samber/lo"
)

func CreateTrie(jsonFields []JsonFieldVar) *Trie {
	trie := NewTrie()
	for _, jsonField := range jsonFields {
		trie.Insert(jsonField)
	}
	return trie
}

type Node struct {
	key      string
	isArray  bool
	fields   []JsonFieldVar
	children map[string]*Node
}

func NewNode(key string, isArray bool) *Node {
	return &Node{key: key, isArray: isArray, fields: []JsonFieldVar{}, children: map[string]*Node{}}
}

func (n *Node) AddChild(child *Node) *Node {
	n.children[child.key] = child
	return n
}

type Trie struct {
	root *Node
}

func (t *Trie) Insert(field JsonFieldVar) {
	elements := field.Path.elements

	currentNode := t.root
	if currentNode == nil {
		currentNode = NewNode(elements[0].token.GetLexeme(), elements[0].arrayToken != nil)
		t.root = currentNode
	} else {
		fieldRootMatchesTrieRoot := currentNode.key == elements[0].token.GetLexeme() &&
			currentNode.isArray == (elements[0].arrayToken != nil)
		if !fieldRootMatchesTrieRoot {
			root := fmt.Sprintf("%s%s", currentNode.key, lo.Ternary(currentNode.isArray, "[]", ""))
			input := fmt.Sprintf("%s%s", elements[0].token.GetLexeme(), lo.Ternary(elements[0].arrayToken != nil, "[]", ""))
			panic(fmt.Sprintf("Field root '%s' does not match trie root '%s'", root, input))
		}
	}

	for _, element := range elements[1:] {
		key := element.token.GetLexeme()
		isArray := element.arrayToken != nil

		node, ok := currentNode.children[key]
		if !ok {
			currentNode.AddChild(NewNode(key, isArray))
		} else {
			if node.isArray != isArray {
				panic(fmt.Sprintf("Field '%s' isArray value does not match existing trie isArray value", key))
			}
		}

		currentNode = currentNode.children[key]
	}

	currentNode.fields = append(currentNode.fields, field)
}

func NewTrie() *Trie {
	return &Trie{}
}

// ---

func TraverseTrie(data interface{}, trie *Trie) {
	traverse(data, trie.root)
}

func traverse(data interface{}, node *Node) {
	for _, field := range node.fields {
		field.AddMatch(fmt.Sprintf("%v", data)) // todo is this the best way?
	}

	if node.isArray {
		switch data.(type) { // todo switch to if statement?
		case []interface{}:
			dataArray := data.([]interface{})
			for _, dataChild := range dataArray {
				traverse(dataChild, node)
			}
			return
		}
	}

	switch data.(type) {
	case string:
	case int:
	case float32:
	case float64:
	case bool:
	case nil:
		if len(node.children) == 0 {
			return
		} else {
			panic(fmt.Sprintf("Hit leaf in data, unexpected for non-leaf node '%v': %v", node, data))
		}
	case []interface{}:
		panic(fmt.Sprintf("Hit array data, but node not marked as array '%v': %v", node, data))
	case map[string]interface{}:
		dataMap := data.(map[string]interface{})
		for childKey, child := range node.children {
			if value, ok := dataMap[childKey]; ok {
				traverse(value, child)
			} else {
				panic(fmt.Sprintf("Expected key '%s' but was not present", childKey))
			}
		}
	default:
		panic(fmt.Sprintf("Expected map for non-array node '%v': %v", node, data))
	}
}
