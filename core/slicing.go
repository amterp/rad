package core

// CalculateCorrectedIndex 'corrects' negative indices into their positive equivalents
func CalculateCorrectedIndex(rawIdx, length int64, clamp bool) int64 {
	idx := rawIdx
	if rawIdx < 0 {
		idx = rawIdx + length
	}

	if clamp {
		if idx < 0 {
			idx = 0
		} else if idx > length {
			idx = length
		}
	}

	return idx
}
