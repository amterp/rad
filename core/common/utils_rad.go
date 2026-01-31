package com

import "strings"

func Truncate(str string, maxLen int64) string {
	runes := []rune(str)
	if TerminalIsUtf8 {
		return string(runes[:maxLen-1]) + "…"
	} else {
		return string(runes[:maxLen-3]) + "..."
	}
}

func Reverse(str string) string {
	runes := []rune(str)
	var builder strings.Builder
	builder.Grow(len(str))
	for i := len(runes) - 1; i >= 0; i-- {
		builder.WriteRune(runes[i])
	}
	return builder.String()
}
