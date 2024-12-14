package core

import "fmt"

func (i *MainInterpreter) VisitVarPathExpr(path VarPath) interface{} {
	val := path.Collection.Accept(i)
	for _, key := range path.Keys {
		val = i.extract(val, key)
	}
	return val
}

// ------------------------------

type VarPathLeafVisitor interface {
	AcceptStrSlice(str RslString, start int64, end int64) RslString
	AcceptListSlice(list []interface{}, start int64, end int64) []interface{}
	AcceptMapElement(m RslMap, key RslString) RslMap
}

type VarPathLeafDeleter struct{}

func (v VarPathLeafDeleter) AcceptStrSlice(str RslString, start int64, end int64) RslString {
	return str.Delete(start, end)
}

func (v VarPathLeafDeleter) AcceptListSlice(list []interface{}, start int64, end int64) []interface{} {
	return append(list[:start], list[end:]...)
}

func (v VarPathLeafDeleter) AcceptMapElement(m RslMap, key RslString) RslMap {
	m.Delete(key)
	return m
}

// --

type VarPathLeafSetter struct {
	Val interface{}
}

func (v VarPathLeafSetter) AcceptStrSlice(str RslString, start int64, end int64) RslString {
	return str.Replace(start, end, ToPrintable(v.Val))
}

func (v VarPathLeafSetter) AcceptListSlice(list []interface{}, start int64, end int64) []interface{} {
	return append(append(list[:start], v.Val), list[end:]...)
}

func (v VarPathLeafSetter) AcceptMapElement(m RslMap, key RslString) RslMap {
	m.Set(key, v.Val)
	return m
}

// ------------------------------

func (i *MainInterpreter) traverseVarPath(tkn Token, col interface{}, keys []CollectionKey, visitor VarPathLeafVisitor) interface{} {
	if len(keys) == 0 {
		return col
	}

	key := keys[0]
	switch coerced := col.(type) {
	case []interface{}:
		if len(keys) == 1 {
			// end of the line, visit whatever we're pointing at
			if key.IsSlice {
				start, end := i.resolveStartEnd(key, int64(len(coerced)))
				return visitor.AcceptListSlice(coerced, start, end)
			} else {
				idx := i.resolveLookupIdx(key, int64(len(coerced)))
				return visitor.AcceptListSlice(coerced, idx, idx+1)
			}
		} else {
			// we want to visit something deeper in the array, recurse
			if key.IsSlice {
				i.error(key.Opener, fmt.Sprintf("Cannot have an intermediary slice: %v", key))
			}

			idx, _ := i.resolveStartEnd(key, int64(len(coerced)))
			coerced[idx] = i.traverseVarPath(tkn, coerced[idx], keys[1:], visitor)
			return coerced
		}
	case RslMap:
		if key.IsSlice {
			i.error(key.Opener, fmt.Sprintf("Cannot slice a map: %v", key))
		}

		keyVal := (*key.Start).Accept(i)
		keyStr, ok := keyVal.(RslString)

		if !ok {
			// todo still unsure about this string constraint
			i.error(tkn, fmt.Sprintf("Map key must be a string, was %v (%v)", TypeAsString(key), key))
		}

		if len(keys) == 1 {
			// end of the line, visit whatever we're pointing at
			return visitor.AcceptMapElement(coerced, keyStr)
		} else {
			// we want to visit something deeper in the map, recurse
			val, exists := coerced.Get(keyStr)
			if !exists {
				i.error(tkn, fmt.Sprintf("Map key %q not found", keyStr))
			}
			coerced.Set(keyStr, i.traverseVarPath(tkn, val, keys[1:], visitor))
			return coerced
		}
	case RslString:
		if len(keys) == 1 {
			// end of the line, visit whatever we're pointing at
			if key.IsSlice {
				start, end := i.resolveStartEnd(key, coerced.Len())
				return visitor.AcceptStrSlice(coerced, start, end)
			} else {
				idx := i.resolveLookupIdx(key, coerced.Len())
				return visitor.AcceptStrSlice(coerced, idx, idx+1)
			}
		} else {
			i.error(keys[1].Opener, fmt.Sprintf("Cannot delete from string slice: %v", keys[1]))
			panic(UNREACHABLE)
		}
	default:
		i.error(tkn, fmt.Sprintf("Cannot delete from a %v", TypeAsString(col)))
		panic(UNREACHABLE)
	}
}
