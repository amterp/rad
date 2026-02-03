package core

import "testing"

func TestFuzzyMatchFold(t *testing.T) {
	tests := []struct {
		source   string
		target   string
		expected bool
	}{
		// Basic matches
		{"", "", true},
		{"", "abc", true},
		{"abc", "abc", true},
		{"abc", "ABC", true},
		{"ABC", "abc", true},

		// Fuzzy matches - characters in order
		{"abc", "aXbXc", true},
		{"abc", "AbraCadabra", true},
		{"fzs", "fuzzysearch", true},
		{"fzs", "FuzzySearch", true},

		// Non-matches
		{"abc", "ab", false},   // source longer than target
		{"abc", "acb", false},  // wrong order
		{"adc", "abcd", false}, // d comes before c
		{"xyz", "abc", false},  // no match at all

		// Edge cases
		{"a", "a", true},
		{"a", "A", true},
		{"a", "bab", true},
		{"aa", "a", false}, // need two a's, only have one
		{"aa", "aXa", true},
	}

	for _, tt := range tests {
		t.Run(tt.source+"_in_"+tt.target, func(t *testing.T) {
			result := FuzzyMatchFold(tt.source, tt.target)
			if result != tt.expected {
				t.Errorf("FuzzyMatchFold(%q, %q) = %v, want %v",
					tt.source, tt.target, result, tt.expected)
			}
		})
	}
}
