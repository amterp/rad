package rl

import (
	ts "github.com/tree-sitter/go-tree-sitter"
)

func NewTypingFnT(fnNode *ts.Node, src string) *TypingFnT {
	paramNodes := GetChildren(fnNode, F_PARAM)
	returnTypeNode := GetChild(fnNode, F_RETURN_TYPE)
	params := make([]TypingFnParam, 0)
	namedOnly := false
	for _, paramNode := range paramNodes {
		nameNode := GetChild(&paramNode, F_NAME)
		typeNode := GetChild(&paramNode, F_TYPE)
		varArgMarkerNode := GetChild(&paramNode, F_VARARG_MARKER)
		optionalNode := GetChild(&paramNode, F_OPTIONAL)
		defaultNode := GetChild(&paramNode, F_DEFAULT)

		if varArgMarkerNode != nil && nameNode == nil {
			// syntax: , *,
			namedOnly = true
			continue
		}

		typingParam := TypingFnParam{
			Name:      GetSrc(nameNode, src),
			NamedOnly: namedOnly,
		}

		var typing TypingT
		if typeNode != nil {
			typing = resolveTyping(typeNode)
			typingParam.Type = &typing
		} else if optionalNode != nil {
			typing = NewOptionalType(NewAnyType())
			typingParam.Type = &typing
			typingParam.IsOptional = true
		}

		if defaultNode != nil {
			typingParam.IsOptional = true
			typingParam.Default = defaultNode
		}

		params = append(params, typingParam)
	}

	typingFn := &TypingFnT{
		Params: params,
	}
	if returnTypeNode != nil {
		returnType := resolveTyping(returnTypeNode)
		typingFn.ReturnT = &returnType
	}
	return typingFn
}

// input node expected to be kind 'fn_param_or_return_type'
func resolveTyping(node *ts.Node) TypingT {
	leafNodes := GetChildren(node, F_LEAF_TYPE)

	if len(leafNodes) == 0 {
		return NewAnyType()
	}

	leafTypes := make([]TypingT, 0)
	for _, leafNode := range leafNodes {
		varargNode := GetChild(&leafNode, F_VARARG_MARKER)
		typeNode := GetChild(&leafNode, F_TYPE)
		optionalNode := GetChild(&leafNode, F_OPTIONAL)

		var typing TypingT
		switch typeNode.Kind() {
		case K_STRING_TYPE:
			typing = NewStrType()
		case K_INT_TYPE:
			typing = NewIntType()
		case K_FLOAT_TYPE:
			typing = NewFloatType()
		case K_BOOL_TYPE:
			typing = NewBoolType()
		case K_STRING_LIST_TYPE:
			typing = NewListType(NewVarArgType(NewStrType()))
		case K_INT_LIST_TYPE:
			typing = NewListType(NewVarArgType(NewIntType()))
		case K_FLOAT_LIST_TYPE:
			typing = NewListType(NewVarArgType(NewFloatType()))
		case K_BOOL_LIST_TYPE:
			typing = NewListType(NewVarArgType(NewBoolType()))
		case K_LIST_TYPE:
			anyNode := GetChild(typeNode, F_ANY)
			if anyNode != nil {
				return NewAnyListType()
			}
			listTypeNodes := GetChildren(typeNode, F_TYPE)
			if len(listTypeNodes) > 0 {
				typings := make([]TypingT, 0)
				for _, listTypeNode := range listTypeNodes {
					typing = resolveTyping(&listTypeNode)
					typings = append(typings, typing)
				}
				return NewListType(typings...)
			}
			enumStrNodes := GetChildren(typeNode, F_ENUM)
			strNodes := make([]*ts.Node, 0)
			for _, enumStrNode := range enumStrNodes {
				strNodes = append(strNodes, &enumStrNode)
			}
			return NewStrEnumType(strNodes...)
		case K_NUM_TYPE:
			typing = NewNumType() // TODO remove
		case K_ANY_TYPE:
			typing = NewAnyType()
		case K_FN_TYPE:
			typing = NewAnyType() // TODO
		case K_MAP_TYPE:
			anyNode := GetChild(typeNode, F_ANY)
			if anyNode != nil {
				return NewAnyMapType()
			}
			entryNodes := GetChildren(typeNode, F_NAMED_ENTRY)
			if len(entryNodes) > 0 {
				keyValues := make(map[MapNamedKey]TypingT)
				for _, entryNode := range entryNodes {
					keyNode := GetChild(&entryNode, F_KEY_NAME)
					keyOptionalNode := GetChild(&entryNode, F_OPTIONAL)
					valueNode := GetChild(&entryNode, F_VALUE_TYPE)
					valueTyping := resolveTyping(valueNode)

					key := NewMapNamedKey(keyNode, keyOptionalNode != nil)
					keyValues[key] = valueTyping
				}
				return NewStructType(keyValues)
			}

			keyTypeNode := GetChild(typeNode, F_KEY_TYPE)
			valueTypeNode := GetChild(typeNode, F_VALUE_TYPE)
			keyTyping := resolveTyping(keyTypeNode)
			valueTyping := resolveTyping(valueTypeNode)
			return NewMapType(keyTyping, valueTyping)
		case K_ERROR_TYPE:
			typing = NewErrorType()
		case K_VOID_TYPE:
			typing = NewVoidType()
		default:
			panic("unknown type node kind: " + typeNode.Kind())
		}
		if optionalNode != nil {
			typing = NewOptionalType(typing)
		}
		if varargNode != nil {
			typing = NewVarArgType(typing)
		}

		leafTypes = append(leafTypes, typing)
	}

	if len(leafTypes) == 1 {
		return leafTypes[0]
	}

	return NewUnionType(leafTypes...)
}
