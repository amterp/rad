package rl

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

// fieldIds caches field name -> numeric ID mappings, eliminating repeated
// CString conversions through the CGo boundary. Initialized once via InitFieldIds.
var fieldIds map[string]uint16

// InitFieldIds pre-computes numeric field IDs for all known field names.
// Must be called once after parsing, before any hot-path GetChildren/GetChild calls.
// Before initialization, GetChildren/GetChild fall back to the original tree-sitter
// string-based lookups (used during init-time signature parsing).
func InitFieldIds(node *ts.Node) {
	lang := node.Language()
	fieldIds = make(map[string]uint16, len(allFieldNames))
	for _, name := range allFieldNames {
		fieldIds[name] = lang.FieldIdForName(name)
	}
}

// GetChildren returns all children of node with the given field name,
// using a reusable cursor to avoid per-call CGo cursor allocation.
//
// Falls back to ChildrenByFieldName if field IDs haven't been initialized
// (e.g. during init-time signature parsing).
func GetChildren(node *ts.Node, fieldName string, cursor *ts.TreeCursor) []ts.Node {
	if fieldIds == nil {
		return node.ChildrenByFieldName(fieldName, cursor)
	}
	fieldId := fieldIds[fieldName]
	if fieldId == 0 {
		return nil
	}
	cursor.Reset(*node)
	if !cursor.GotoFirstChild() {
		return nil
	}
	var result []ts.Node
	for {
		if cursor.FieldId() == fieldId {
			n := cursor.Node()
			result = append(result, *n)
		}
		if !cursor.GotoNextSibling() {
			break
		}
	}
	return result
}

// GetChild returns the first child of node with the given field name,
// using cached field IDs to avoid CString conversion.
//
// Falls back to ChildByFieldName if field IDs haven't been initialized.
func GetChild(node *ts.Node, fieldName string) *ts.Node {
	if fieldIds == nil {
		return node.ChildByFieldName(fieldName)
	}
	return node.ChildByFieldId(fieldIds[fieldName])
}

func GetSrc(node *ts.Node, src string) string {
	return src[node.StartByte():node.EndByte()]
}

// allFieldNames lists every F_* constant for field ID pre-computation.
var allFieldNames = []string{
	F_LEFT, F_LEFTS, F_RIGHT, F_ROOT, F_INDEXING, F_INDEX,
	F_LIST_ENTRY, F_FUNC, F_ARG, F_NAMED_ARG, F_OP, F_CONTENTS,
	F_EXPR, F_FORMAT, F_THOUSANDS_SEPARATOR, F_ALIGNMENT, F_PADDING,
	F_PRECISION, F_MAP_ENTRY, F_KEY, F_VALUE, F_ALT, F_CONDITION,
	F_STMT, F_KEYWORD, F_START, F_END, F_SHELL_CMD, F_MODIFIER,
	F_QUIET_MOD, F_CONFIRM_MOD, F_COMMAND, F_SEGMENT, F_SOURCE,
	F_RAD_TYPE, F_IDENTIFIER, F_SPECIFIER, F_MOD_STMT, F_COLOR,
	F_REGEX, F_LAMBDA, F_TRUE_BRANCH, F_FALSE_BRANCH, F_NAME,
	F_FIRST, F_SECOND, F_DISCRIMINANT, F_CASE, F_CASE_KEY,
	F_DEFAULT, F_YIELD_STMT, F_RETURN_STMT, F_DELEGATE, F_PARAM,
	F_METADATA_ENTRY, F_BLOCK_COLON, F_CATCH, F_TYPE, F_RETURN_TYPE,
	F_VARARG_MARKER, F_VARIADIC_MARKER, F_OPTIONAL, F_LEAF_TYPE,
	F_ANY, F_ENUM, F_NAMED_ENTRY, F_KEY_NAME, F_KEY_TYPE,
	F_VALUE_TYPE, F_LIST, F_NORMAL_PARAM, F_NAMED_ONLY_PARAM,
	F_VARARG_PARAM, F_CALLBACK_IDENTIFIER, F_CALLBACK_LAMBDA,
	F_DESCRIPTION, F_CALLS, F_CONTEXT,
}
