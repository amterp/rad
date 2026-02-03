package core

import (
	"strings"
	"unicode/utf8"
)

// FuzzyMatchFold returns true if each character in source can be found in
// target in order, using case-insensitive matching. This is a simplified
// fuzzy search - not Levenshtein distance.
//
// For example:
//   - FuzzyMatchFold("abc", "AbraCadabra") -> true (a...b...c found in order)
//   - FuzzyMatchFold("adc", "abcd") -> false (d comes before c in target)
func FuzzyMatchFold(source, target string) bool {
	source = strings.ToLower(source)
	target = strings.ToLower(target)

	if len(source) > len(target) {
		return false
	}

	if source == target {
		return true
	}

	for _, r1 := range source {
		found := false
		for i, r2 := range target {
			if r1 == r2 {
				target = target[i+utf8.RuneLen(r2):]
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	return true
}
