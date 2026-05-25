package rl

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

// K_FN_NAMED expected as node
func NewTypingFnT(fnNode *ts.Node, src string) *TypingFnT {
	nameNode := GetChild(fnNode, F_NAME)
	normalParamNodes := GetChildren(fnNode, F_NORMAL_PARAM)
	varargNode := GetChild(fnNode, F_VARARG_PARAM)
	namedOnlyNodes := GetChildren(fnNode, F_NAMED_ONLY_PARAM)
	returnTypeNode := GetChild(fnNode, F_RETURN_TYPE)

	params := make([]TypingFnParam, 0)
	resolveParams(&params, src, normalParamNodes, false, false)
	if varargNode != nil {
		resolveParams(&params, src, []ts.Node{*varargNode}, true, false)
	}
	resolveParams(&params, src, namedOnlyNodes, false, true)

	typingFn := &TypingFnT{
		Params: params,
	}

	if nameNode != nil {
		typingFn.FnName = GetSrc(nameNode, src)
	}

	if returnTypeNode != nil {
		returnType := resolveTyping(returnTypeNode, src)
		typingFn.ReturnT = &returnType
	}
	return typingFn
}

func resolveParams(resolvedParams *[]TypingFnParam, src string, paramNodes []ts.Node, isVariadic, isNamedOnly bool) {
	for _, paramNode := range paramNodes {
		nameNode := GetChild(&paramNode, F_NAME)
		typeNode := GetChild(&paramNode, F_TYPE)
		optionalNode := GetChild(&paramNode, F_OPTIONAL)
		defaultNode := GetChild(&paramNode, F_DEFAULT)

		typingParam := TypingFnParam{
			Name:       GetSrc(nameNode, src),
			NameSpan:   spanFromNode(nameNode),
			IsVariadic: isVariadic,
			NamedOnly:  isNamedOnly,
		}

		var typing TypingT
		if typeNode != nil {
			typing = resolveTyping(typeNode, src)
			typingParam.Type = &typing
		} else if optionalNode != nil {
			typing = NewOptionalType(NewAnyType())
			typingParam.Type = &typing
			typingParam.IsOptional = true
		}

		if defaultNode != nil {
			typingParam.IsOptional = true
			typingParam.Default = NewRadNode(defaultNode, src)
		}

		*resolvedParams = append(*resolvedParams, typingParam)
	}
}

// ResolveTyping is the public entry point to resolveTyping for callers
// outside the rl package (e.g. the converter wiring typed locals).
// Accepts a CST node of kind fn_param_or_return_type and returns the
// corresponding static TypingT.
func ResolveTyping(node *ts.Node, src string) TypingT {
	return resolveTyping(node, src)
}

// input node expected to be kind 'fn_param_or_return_type'
func resolveTyping(node *ts.Node, src string) TypingT {
	leafNodes := GetChildren(node, F_LEAF_TYPE)

	if len(leafNodes) == 0 {
		return NewAnyType()
	}

	leafTypes := make([]TypingT, 0)
	for _, leafNode := range leafNodes {
		listNodes := GetChildren(&leafNode, F_LIST)
		optionalNode := GetChild(&leafNode, F_OPTIONAL)

		// Parenthesized group form: `(int|str)[]?`. Recurse to resolve
		// the inner union, then apply the outer list / optional
		// modifiers below. The group field replaces the type field for
		// this shape; checked first so the type-kind switch can assume
		// a non-nil typeNode.
		var typing TypingT
		groupNode := GetChild(&leafNode, F_GROUP)
		if groupNode != nil {
			typing = resolveTyping(groupNode, src)
			for range listNodes {
				typing = NewListType(typing)
			}
			if optionalNode != nil {
				typing = NewOptionalType(typing)
			}
			leafTypes = append(leafTypes, typing)
			continue
		}

		typeNode := GetChild(&leafNode, F_TYPE)

		switch typeNode.Kind() {
		case K_STRING_TYPE:
			typing = NewStrType()
		case K_INT_TYPE:
			typing = NewIntType()
		case K_FLOAT_TYPE:
			typing = NewFloatType()
		case K_BOOL_TYPE:
			typing = NewBoolType()
		case K_LIST_TYPE:
			anyNode := GetChild(typeNode, F_ANY)
			if anyNode != nil {
				typing = NewAnyListType()
				break
			}
			listTypeNodes := GetChildren(typeNode, F_TYPE)
			if len(listTypeNodes) > 0 {
				typings := make([]TypingT, 0)
				for _, listTypeNode := range listTypeNodes {
					typing = resolveTyping(&listTypeNode, src)
					typings = append(typings, typing)
				}
				typing = NewTupleType(typings...)
				break
			}
			enumStrNodes := GetChildren(typeNode, F_ENUM)
			values := make([]string, 0, len(enumStrNodes))
			for _, enumStrNode := range enumStrNodes {
				// Extract string value by stripping quotes
				raw := GetSrc(&enumStrNode, src)
				if len(raw) >= 2 {
					raw = raw[1 : len(raw)-1]
				}
				values = append(values, raw)
			}
			typing = NewStrEnumType(values...)
		case K_ANY_TYPE:
			typing = NewAnyType()
		case K_FN_TYPE:
			typing = newTypingFnTFromType(typeNode, src)
		case K_MAP_TYPE:
			anyNode := GetChild(typeNode, F_ANY)
			if anyNode != nil {
				typing = NewAnyMapType()
				break
			}
			entryNodes := GetChildren(typeNode, F_NAMED_ENTRY)
			if len(entryNodes) > 0 {
				keyValues := make(map[MapNamedKey]TypingT)
				for _, entryNode := range entryNodes {
					keyNode := GetChild(&entryNode, F_KEY_NAME)
					keyOptionalNode := GetChild(&entryNode, F_OPTIONAL)
					valueNode := GetChild(&entryNode, F_VALUE_TYPE)
					valueTyping := resolveTyping(valueNode, src)

					// Extract the key name string by stripping quotes
					rawKey := GetSrc(keyNode, src)
					if len(rawKey) >= 2 {
						rawKey = rawKey[1 : len(rawKey)-1]
					}
					key := NewMapNamedKey(rawKey, keyOptionalNode != nil)
					keyValues[key] = valueTyping
				}
				typing = NewStructType(keyValues)
				break
			}

			keyTypeNode := GetChild(typeNode, F_KEY_TYPE)
			valueTypeNode := GetChild(typeNode, F_VALUE_TYPE)
			keyTyping := resolveTyping(keyTypeNode, src)
			valueTyping := resolveTyping(valueTypeNode, src)
			typing = NewMapType(keyTyping, valueTyping)
		case K_ERROR_TYPE:
			typing = NewErrorType()
		case K_VOID_TYPE:
			typing = NewVoidType()
		default:
			panic("unknown type node kind: " + typeNode.Kind())
		}
		// Wrap once per "[]" suffix so `int[][]` becomes NewListType(NewListType(IntT)).
		for range listNodes {
			typing = NewListType(typing)
		}
		if optionalNode != nil {
			typing = NewOptionalType(typing)
		}

		leafTypes = append(leafTypes, typing)
	}

	if len(leafTypes) == 1 {
		return leafTypes[0]
	}

	return NewUnionType(leafTypes...)
}

// newTypingFnTFromType builds a TypingFnT from a K_FN_TYPE node (the type
// annotation `fn(int, str) -> bool`). The shape differs from K_FN_NAMED:
// no function name, no default values, no named-only params - just typed
// positional params and an optional return type.
func newTypingFnTFromType(fnTypeNode *ts.Node, src string) *TypingFnT {
	paramNodes := GetChildren(fnTypeNode, F_PARAM)
	returnTypeNode := GetChild(fnTypeNode, F_RETURN_TYPE)

	params := make([]TypingFnParam, 0, len(paramNodes))
	for _, paramNode := range paramNodes {
		typing := resolveTyping(&paramNode, src)
		params = append(params, TypingFnParam{Type: &typing})
	}

	typingFn := &TypingFnT{Params: params}
	if returnTypeNode != nil {
		returnType := resolveTyping(returnTypeNode, src)
		typingFn.ReturnT = &returnType
	}
	return typingFn
}
