package core

import (
	"fmt"
)

func CreateTrie(i *MainInterpreter, radToken Token, jsonFields []JsonFieldVar) *Trie {
	trie := &Trie{i: i, radToken: radToken}
	for _, jsonField := range jsonFields {
		trie.Insert(jsonField)
	}
	return trie
}

type Node struct {
	radToken           Token
	key                string
	shouldIterateArray bool
	idx                *int64
	// json variables which terminate at this node, and therefore need to capture the data at this level
	fields   []JsonFieldVar
	children map[string]*Node
}

func NewNode(radToken Token, key string, shouldIterateArray bool, idx *int64) *Node {
	return &Node{
		radToken:           radToken,
		key:                key,
		shouldIterateArray: shouldIterateArray,
		idx:                idx,
		fields:             []JsonFieldVar{},
		children:           map[string]*Node{},
	}
}

func (n *Node) AddChild(child *Node) *Node {
	n.children[child.key] = child
	return n
}

type Trie struct {
	i        *MainInterpreter
	radToken Token
	root     *Node
}

func (t *Trie) Insert(field JsonFieldVar) {
	elements := field.Path.Elements

	currentNode := t.root
	if currentNode == nil {
		rootElement := elements[0]
		idx := tryEvalIdx(t.i, rootElement)
		currentNode = NewNode(t.radToken, rootElement.Identifier.GetLexeme(), rootElement.ShouldIterateArray(), idx)
		t.root = currentNode
	}

	for _, element := range elements[1:] {
		key := element.Identifier.GetLexeme()
		_, ok := currentNode.children[key]
		if !ok {
			idx := tryEvalIdx(t.i, element)
			currentNode.AddChild(NewNode(t.radToken, key, element.ShouldIterateArray(), idx))
		}

		currentNode = currentNode.children[key]
	}

	currentNode.fields = append(currentNode.fields, field)
}

func tryEvalIdx(i *MainInterpreter, elem JsonPathElement) *int64 {
	var idx *int64
	if elem.Index != nil {
		val := (*elem.Index).Accept(i)
		coercedVal, ok := val.(int64)
		if !ok {
			i.error(elem.Identifier, fmt.Sprintf("Expected int for array index, got %v (%s)", val, TypeAsString(val)))
		}
		idx = &coercedVal
	}
	return idx
}

// ---

func (t *Trie) TraverseTrie(data interface{}) {
	t.traverse(data, t.root, nil)
}

func (t *Trie) traverse(data interface{}, node *Node, keyToCaptureInstead interface{}) captureStats {
	capStats := captureStats{
		captures: 0,
		wasLeaf:  false,
	}

	if node.shouldIterateArray {
		dataArray, ok := data.([]interface{})
		if !ok {
			// todo feels like we should error here, but in practice does not work, investigate
			//RP.TokenErrorExit(node.radToken, fmt.Sprintf("Expected array for array node '%v': %v\n", node, data))
		} else {
			if node.idx != nil {
				capStats = capStats.add(t.traverse(dataArray[*node.idx], node, nil))
			} else {
				for _, dataChild := range dataArray {
					capStats = capStats.add(t.traverse(dataChild, node, nil))
				}
			}
			t.capture(data, node, keyToCaptureInstead, capStats.captures)
			return capStats
		}
	}

	switch coerced := data.(type) {
	case string, int, float32, float64, bool, nil:
		// leaf
		if len(node.children) == 0 {
			capStats = captureStats{1, true}
		} else {
			RP.TokenErrorExit(node.radToken, fmt.Sprintf("Hit leaf in data, unexpected for non-leaf node '%v': %v\n", node, data))
		}
	case []interface{}:
		if len(node.children) == 0 {
			capStats = captureStats{1, true}
		} else {
			RP.TokenErrorExit(node.radToken, fmt.Sprintf("Hit array data, but node not marked as array '%v': %v", node, data))
		}
	case map[string]interface{}:
		dataMap := coerced
		for childKey, child := range node.children {
			if childKey == WILDCARD {
				// wildcard match, traverse all children
				// get list of sorted keys to iterate through, for deterministic order
				// todo: at this point this is a concession -- we should be traversing in the original order of the json
				//  but the json.Unmarshalling we're doing loses us the original ordering for maps. will need to change
				//  how we get the json if we want to change that (we almost certainly do, this is idiosyncratic behavior)
				sortedKeys := SortedKeys(dataMap)
				for _, key := range sortedKeys {
					capStats = capStats.add(t.traverse(dataMap[key], child, key))
				}
			} else if value, ok := dataMap[childKey]; ok {
				capStats = capStats.add(t.traverse(value, child, nil))
			} else {
				RP.TokenErrorExit(node.radToken, fmt.Sprintf("Expected key '%s' but was not present\n", childKey))
			}
		}
		if len(node.fields) > 0 && node.key != WILDCARD {
			// we're at a dictionary node and being asked to capture. let's capture the node as JSON
			// max: we want to capture at least once, but if we've captured from children nodes, we want to capture
			// that many
			t.capture(dataMap, node, keyToCaptureInstead, max(capStats.captures, 1))
		}
	default:
		RP.TokenErrorExit(node.radToken, fmt.Sprintf("Expected map for non-array node '%v': %v\n", node, data))
	}

	t.capture(data, node, keyToCaptureInstead, capStats.captures)
	return capStats
}

func (t *Trie) capture(data interface{}, node *Node, keyToCaptureInstead interface{}, captures int) {
	for i := 0; i < captures; i++ {
		if node.idx != nil {
			if _, isArray := data.([]interface{}); !isArray {
				for _, field := range node.fields {
					field.AddMatch(data)
				}
			}
		} else if keyToCaptureInstead == nil && node.key != WILDCARD {
			for _, field := range node.fields {
				field.AddMatch(data)
			}
		} else if keyToCaptureInstead != nil {
			for _, field := range node.fields {
				field.AddMatch(keyToCaptureInstead)
			}
		}
	}
}

type captureStats struct {
	// aka rows
	captures int
	wasLeaf  bool
}

func (c captureStats) add(other captureStats) captureStats {
	wasLeaf := false
	if other.wasLeaf {
		wasLeaf = true
	}
	if wasLeaf {
		c.captures = 1
	} else {
		c.captures += other.captures
	}
	return c
}
