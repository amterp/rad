package core

import "fmt"

func (i *MainInterpreter) extract(col interface{}, colKey CollectionKey) interface{} {
	if col == nil {
		i.error(colKey.Opener, "Cannot slice a nil value")
	}

	if colKey.IsSlice {
		return i.sliceAccess(col, colKey)
	} else {
		if colKey.End == nil {
			return i.colLookup(col, colKey)
		} else {
			i.error(colKey.Opener, fmt.Sprintf("Bug! Non-slice cannot have an end index: %v", colKey))
			panic(UNREACHABLE)
		}
	}
}

func (i *MainInterpreter) sliceAccess(col interface{}, key CollectionKey) interface{} {
	switch coerced := col.(type) {
	case RslString:
		start, end := i.resolveStartEnd(key, coerced.Len())
		return coerced.Slice(start, end)
	case []interface{}:
		start, end := i.resolveStartEnd(key, int64(len(coerced)))
		return coerced[start:end]
	default:
		i.error(key.Opener, "Slice access must be on a string or array")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) colLookup(col interface{}, key CollectionKey) interface{} {
	switch coerced := col.(type) {
	case RslMap:
		start := *key.Start
		keyVal := start.Accept(i)
		keyStr, ok := keyVal.(RslString)
		if !ok {
			i.error(key.Opener, fmt.Sprintf("Map key must be a string, was %T (%v)", keyVal, keyVal))
		}
		if val, exists := coerced.Get(keyStr); exists {
			return val
		} else {
			i.error(key.Opener, fmt.Sprintf("Key not found: %v", keyStr.Plain()))
			panic(UNREACHABLE)
		}
	case []interface{}:
		idx := i.resolveLookupIdx(key, int64(len(coerced)))
		return coerced[idx]
	case RslString:
		idx := i.resolveLookupIdx(key, coerced.Len())
		return coerced.IndexAt(idx)
	default:
		i.error(key.Opener, "Lookup must be on a map, array, or string")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) resolveLookupIdx(key CollectionKey, len int64) int64 {
	if key.IsSlice {
		i.error(key.Opener, fmt.Sprintf("Cannot perform lookup with a slice key; %v", key))
	}

	keyVal := (*key.Start).Accept(i)
	switch coercedKey := keyVal.(type) {
	case int64:
		adjustedKey := coercedKey
		if adjustedKey < 0 {
			adjustedKey += len
		}
		if adjustedKey < 0 || adjustedKey >= len {
			i.error(key.Opener, fmt.Sprintf("Index out of bounds: %d (length: %d)", coercedKey, len))
		}
		return adjustedKey
	default:
		i.error(key.Opener, "Lookup key must be an int")
		panic(UNREACHABLE)
	}
}

func (i *MainInterpreter) resolveStartEnd(key CollectionKey, len int64) (int64, int64) {
	start := int64(0)
	end := len

	if key.Start != nil {
		start = i.resolveSliceIndex(key.Opener, *key.Start, len, true)
	}
	if key.End != nil {
		end = i.resolveSliceIndex(key.Opener, *key.End, len, false)
	}

	if start > end {
		start = end
	}

	return start, end
}

func (i *MainInterpreter) resolveSliceIndex(token Token, expr Expr, len int64, isStart bool) int64 {
	index := expr.Accept(i)
	rawIdx, ok := index.(int64)
	if !ok {
		i.error(token, fmt.Sprintf("Slice index must be an int, was %T (%v)", index, index))
	}

	var idx = rawIdx
	if rawIdx < 0 {
		idx = rawIdx + len
	}

	if isStart {
		if idx < 0 {
			// the start index is still negative, so we'll slice from the beginning
			idx = 0
		}
		if idx > len {
			// the start index is greater than the length of the list, so we'll slice to the end
			idx = len
		}
	} else {
		if idx > len {
			// the end index is greater than the length of the list, so we'll slice to the end
			idx = len
		}
		if idx < 0 {
			// the end index is still negative, so we'll slice from the end
			idx = 0
		}
	}

	return idx
}
