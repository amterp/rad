package core

import "fmt"

func (i *MainInterpreter) extract(col interface{}, colKey CollectionKey) interface{} {
	return nil // DELETE
}

func (i *MainInterpreter) sliceAccess(col interface{}, key CollectionKey) interface{} {
	return nil // DELETE
}

func (i *MainInterpreter) colLookup(col interface{}, key CollectionKey) interface{} {
	return nil // DELETE
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
	return 0, 0 // DELETE
}

func (i *MainInterpreter) resolveSliceIndex(token Token, expr Expr, len int64, isStart bool) int64 {
	return 0 // DELETE
}
