package com

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/dustin/go-humanize/english"
)

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

func SortedKeys[T any](m map[string]T) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Simple len(str) call counts bytes, not runes, so e.g. emojis gets counted as multiple characters
func StrLen(str string) int {
	return utf8.RuneCountInString(str)
}

func IsBlank(str string) bool {
	return strings.TrimSpace(str) == ""
}

func Memoize[T any](f func() T) func() T {
	var once sync.Once
	var result T

	return func() T {
		once.Do(func() {
			result = f()
		})
		return result
	}
}

// follows some basic rules of english, use PluralizeCustom to override the plural.
func Pluralize(count int, singular string) string {
	return english.Plural(count, singular, "")
}

func PluralizeCustom(count int, singular string, plural string) string {
	return english.Plural(count, singular, plural)
}

func NumIsAre(values int) string {
	if values == 1 {
		return fmt.Sprintf("%d is", values)
	} else {
		return fmt.Sprintf("%d are", values)
	}
}

func Int64Min(a, b int64) int64 {
	return int64(math.Min(float64(a), float64(b)))
}

func Int64Max(a, b int64) int64 {
	return int64(math.Max(float64(a), float64(b)))
}

func IntMin(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}

func IntMax(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}
