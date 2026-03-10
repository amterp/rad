package core

import (
	"strings"
	"unicode/utf8"
)

// Levenshtein calculates the Levenshtein distance between two strings.
// Used for "did you mean?" suggestions.
func Levenshtein(a, b string) int {
	ra := []rune(a)
	rb := []rune(b)

	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}

	// Create matrix
	matrix := make([][]int, len(ra)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(rb)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(rb); j++ {
		matrix[0][j] = j
	}

	// Fill matrix
	for i := 1; i <= len(ra); i++ {
		for j := 1; j <= len(rb); j++ {
			cost := 0
			if ra[i-1] != rb[j-1] {
				cost = 1
			}
			matrix[i][j] = min(
				matrix[i-1][j]+1,      // deletion
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(ra)][len(rb)]
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
