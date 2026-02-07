package core

import (
	"strings"
	"unicode/utf8"
)

// Levenshtein calculates the Levenshtein distance between two strings.
// Used for "did you mean?" suggestions.
func Levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	// Create matrix
	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

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
