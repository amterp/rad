package core

import "sort"

// this is the best way I can think of to do the 'typed nil' check...
func NotNil[T comparable](val *T, nilProvider func() T) bool {
	if val == nil {
		return false
	}

	if (*val) == nilProvider() {
		return false
	}

	return true
}

func AllNils[T comparable](vals []*T) bool {
	for _, val := range vals {
		if val != nil {
			return false
		}
	}

	return true
}

func SortedKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
