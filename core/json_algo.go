package core

import (
	com "rad/core/common"

	"github.com/samber/lo"
)

func (t *Trie) TraverseTrie(data interface{}) {
	jsonRoot := lo.Values(t.root.children)[0]
	captures := t.traverse(nil, data, jsonRoot)
	for varName, values := range captures.captures {
		t.i.env.SetVar(varName, newRadValue(t.i, t.radKeywordNode, values))
		// todo
		//  - we *always* wrap in an array. some way to encode "expect non-array"?
	}
}

func (t *Trie) traverse(dataKey *string, data interface{}, node *TrieNode) Capture {
	// todo capture whole block if 'json' is one of the fields

	// traverse children, capture values
	captures := make([]Capture, 0)
	for _, child := range lo.Values(node.children) {
		if child.isArrayWildcard {
			// iterate through all elements
			switch coerced := data.(type) {
			case []interface{}:
				for _, elem := range coerced {
					captures = append(captures, t.traverse(nil, elem, child))
				}
			default:
				t.i.errorf(t.radKeywordNode, "Expected array at %s, got %s", child.fullKey, TypeAsString(data))
			}
		} else if child.idx != nil {
			// array index lookup
			switch coerced := data.(type) {
			case []interface{}:
				if int(*child.idx) >= len(coerced) {
					t.i.errorf(t.radKeywordNode, "Index out of bounds at %s: %d", child.fullKey, *child.idx)
				}
				captures = append(captures, t.traverse(nil, coerced[*child.idx], child))
			default:
				t.i.errorf(t.radKeywordNode, "Expected array at %s, got %s", child.fullKey, TypeAsString(data))
			}
		} else if child.key == WILDCARD {
			// wildcard key match
			switch coerced := data.(type) {
			case map[string]interface{}:
				sortedKeys := com.SortedKeys(coerced)
				for _, key := range sortedKeys {
					captures = append(captures, t.traverse(&key, coerced[key], child))
				}
			}
		} else {
			// regular key lookup
			switch coerced := data.(type) {
			case map[string]interface{}:
				childData, ok := coerced[child.key]
				if !ok {
					// todo allow defaulting?
					t.i.errorf(t.radKeywordNode, "Key not found in JSON: "+child.fullKey)
				}
				captures = append(captures, t.traverse(&child.key, childData, child))
			default:
				t.i.errorf(t.radKeywordNode, "Expected map at %s, got %s", child.fullKey, TypeAsString(data))
			}
		}
	}

	capture := t.mergeCaptures(captures, node)

	// capture values at this level

	localCaptures := make(map[string][]interface{})
	for _, field := range node.fields {
		if node.key == WILDCARD {
			if dataKey == nil {
				t.i.errorf(t.radKeywordNode, "Expected data key at %s, got nil", node.fullKey)
			}
			localCaptures[field.Name] = []interface{}{*dataKey}
		} else {
			localCaptures[field.Name] = []interface{}{data}
		}
	}

	localCapture := Capture{node: node, captures: localCaptures}
	capture = t.mergeCapture(capture, localCapture, node)

	return capture
}

func (t *Trie) mergeCaptures(captures []Capture, node *TrieNode) Capture {
	if len(captures) == 0 {
		return Capture{node: node, captures: make(map[string][]interface{})}
	}

	capture := captures[0]
	for _, c := range captures[1:] {
		capture = t.mergeCapture(capture, c, node)
	}

	return capture
}

func (t *Trie) mergeCapture(capture1 Capture, capture2 Capture, node *TrieNode) Capture {
	if len(capture1.captures) == 0 {
		return capture2
	} else if len(capture2.captures) == 0 {
		return capture1
	}

	// check if all columns are the same, if so, append rows
	if len(capture1.captures) == len(capture2.captures) {
		colsAreTheSame := true
		for key, _ := range capture1.captures {
			if _, ok := capture2.captures[key]; !ok {
				colsAreTheSame = false
				break
			}
		}
		if colsAreTheSame {
			for key, values := range capture2.captures {
				capture1.captures[key] = append(capture1.captures[key], values...)
			}
			return capture1
		}
	}

	// check if overlapping columns. if so, error
	for key, _ := range capture1.captures {
		if _, ok := capture2.captures[key]; ok {
			t.i.errorf(
				t.radKeywordNode,
				"Cannot merge captures: %s and %s",
				capture1.node.fullKey,
				capture2.node.fullKey,
			)
		}
	}

	// columns are non-overlapping. if equal number of rows, append columns.
	numRows1 := len(lo.Values(capture1.captures)[0])
	numRows2 := len(lo.Values(capture2.captures)[0])
	if numRows1 == numRows2 {
		for key, values := range capture2.captures {
			capture1.captures[key] = values
		}
		return capture1
	}

	// if one of the numRows is 1, then we'll append its columns and repeat the values to match the others' numRows
	if numRows1 == 1 {
		for key, values := range capture1.captures {
			capture2.captures[key] = make([]interface{}, 0)
			for i := 0; i < numRows2; i++ {
				capture2.captures[key] = append(capture2.captures[key], values[0])
			}
		}
		return capture2
	}

	if numRows2 == 1 {
		for key, values := range capture2.captures {
			capture1.captures[key] = make([]interface{}, 0)
			for i := 0; i < numRows1; i++ {
				capture1.captures[key] = append(capture1.captures[key], values[0])
			}
		}
		return capture1
	}

	// if neither numRows is 1, error
	t.i.errorf(t.radKeywordNode, "Cannot merge captures: %s and %s", capture1.node.fullKey, capture2.node.fullKey)
	panic(UNREACHABLE)
}
