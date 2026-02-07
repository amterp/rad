package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestLevenshtein(t *testing.T) {
	tests := []struct {
		a, b     string
		expected int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "a", 1},
		{"abc", "abc", 0},
		{"abc", "abd", 1},
		{"abc", "ab", 1},
		{"abc", "abcd", 1},
		{"kitten", "sitting", 3},
		{"foobar", "foobaz", 1},
	}

	for _, tc := range tests {
		result := Levenshtein(tc.a, tc.b)
		assert.Equal(t, tc.expected, result, "Levenshtein(%q, %q)", tc.a, tc.b)
	}
}
