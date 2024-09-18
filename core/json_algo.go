package core

import (
	"fmt"
	"github.com/samber/lo"
)

func CreateTrie(radToken Token, jsonFields []JsonFieldVar) *Trie {
	trie := &Trie{radToken: radToken}
	for _, jsonField := range jsonFields {
		trie.Insert(jsonField)
	}
	return trie
}

type Node struct {
	radToken Token
	key      string
	isArray  bool
	fields   []JsonFieldVar
	children map[string]*Node
}

func NewNode(radToken Token, key string, isArray bool) *Node {
	return &Node{
		radToken: radToken,
		key:      key,
		isArray:  isArray,
		fields:   []JsonFieldVar{},
		children: map[string]*Node{},
	}
}

func (n *Node) AddChild(child *Node) *Node {
	n.children[child.key] = child
	return n
}

type Trie struct {
	radToken Token
	root     *Node
}

func (t *Trie) Insert(field JsonFieldVar) {
	elements := field.Path.elements

	currentNode := t.root
	if currentNode == nil {
		currentNode = NewNode(t.radToken, elements[0].token.Literal, elements[0].token.IsArray)
		t.root = currentNode
	} else {
		fieldRootMatchesTrieRoot := currentNode.key == elements[0].token.Literal &&
			currentNode.isArray == elements[0].token.IsArray
		if !fieldRootMatchesTrieRoot {
			root := fmt.Sprintf("%s%s", currentNode.key, lo.Ternary(currentNode.isArray, "[]", ""))
			input := fmt.Sprintf("%s%s", elements[0].token.Literal, lo.Ternary(elements[0].token.IsArray, "[]", ""))
			RP.TokenErrorExit(t.radToken, fmt.Sprintf("Field root '%s' does not match trie root '%s'\n", root, input))
		}
	}

	for _, element := range elements[1:] {
		key := element.token.Literal
		isArray := element.token.IsArray

		node, ok := currentNode.children[key]
		if !ok {
			currentNode.AddChild(NewNode(t.radToken, key, isArray))
		} else {
			if node.isArray != isArray {
				RP.TokenErrorExit(t.radToken, fmt.Sprintf("Field '%s' isArray value does not match existing trie isArray value\n", key))
			}
		}

		currentNode = currentNode.children[key]
	}

	currentNode.fields = append(currentNode.fields, field)
}

// ---

func (t *Trie) TraverseTrie(data interface{}) {
	traverse(data, t.root)
}

func traverse(data interface{}, node *Node) {
	for _, field := range node.fields {
		field.AddMatch(fmt.Sprintf("%v", data)) // todo is this the best way?
	}

	if node.isArray {
		dataArray, ok := data.([]interface{})
		if !ok {
			// todo feels like we should error here, but in practice does not work, investigate
			//RP.TokenErrorExit(node.radToken, fmt.Sprintf("Expected array for array node '%v': %v\n", node, data))
		} else {
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
		}
		RP.TokenErrorExit(node.radToken, fmt.Sprintf("Hit leaf in data, unexpected for non-leaf node '%v': %v\n", node, data))
	case []interface{}:
		if len(node.children) == 0 {
			return
		}
		RP.TokenErrorExit(node.radToken, fmt.Sprintf("Hit array data, but node not marked as array '%v': %v", node, data))
	case map[string]interface{}:
		dataMap := data.(map[string]interface{})
		for childKey, child := range node.children {
			if value, ok := dataMap[childKey]; ok {
				traverse(value, child)
			} else {
				RP.TokenErrorExit(node.radToken, fmt.Sprintf("Expected key '%s' but was not present\n", childKey))
			}
		}
	default:
		RP.TokenErrorExit(node.radToken, fmt.Sprintf("Expected map for non-array node '%v': %v\n", node, data))
	}
}
