package core

import "fmt"

func (i *MainInterpreter) VisitVarPathExpr(path VarPath) interface{} {
	val := path.Collection.Accept(i)
	for _, key := range path.Keys {
		val = i.extract(val, key)
	}
	return val
}

func (i *MainInterpreter) executeDelete(identifier Token, col interface{}, keys []CollectionKey) interface{} {
	if len(keys) == 0 {
		return col
	}

	key := keys[0]
	switch coerced := col.(type) {
	case []interface{}:
		if len(keys) == 1 {
			// end of the line, delete whatever we're pointing at
			if key.IsSlice {
				start, end := i.resolveStartEnd(key, int64(len(coerced)))
				return append(coerced[:start], coerced[end:]...)
			} else {
				idx := i.resolveLookupIdx(key, int64(len(coerced)))
				return append(coerced[:idx], coerced[idx+1:]...)
			}
		} else {
			// we want to delete something deeper in the array, recurse
			if key.IsSlice {
				i.error(key.Opener, fmt.Sprintf("Cannot have an intermediary slice in a delete operation: %v", key))
			}

			idx, _ := i.resolveStartEnd(key, int64(len(coerced)))
			coerced[idx] = i.executeDelete(identifier, coerced[idx], keys[1:])
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
			i.error(identifier, fmt.Sprintf("Map key must be a string, was %v (%v)", TypeAsString(key), key))
		}

		val, exists := coerced.Get(keyStr)
		if !exists {
			i.error(identifier, fmt.Sprintf("Map key %q not found", keyStr))
		}

		if len(keys) == 1 {
			// end of the line, delete whatever we're pointing at
			coerced.Delete(keyStr)
			return coerced
		} else {
			// we want to delete something deeper in the map, recurse
			coerced.Set(keyStr, i.executeDelete(identifier, val, keys[1:]))
			return coerced
		}
	case RslString:
		if len(keys) == 1 {
			// end of the line, delete whatever we're pointing at
			if key.IsSlice {
				start, end := i.resolveStartEnd(key, coerced.Len())
				return coerced.Delete(start, end)
			} else {
				idx := i.resolveLookupIdx(key, coerced.Len())
				return coerced.Delete(idx, idx+1)
			}
		} else {
			i.error(keys[1].Opener, fmt.Sprintf("Cannot delete from string slice: %v", keys[1]))
			panic(UNREACHABLE)
		}
	default:
		i.error(identifier, fmt.Sprintf("Cannot delete from a %v", TypeAsString(col)))
		panic(UNREACHABLE)
	}
}
