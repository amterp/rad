package rts

import (
	"strings"
)

// NormalizeIndentedText removes common leading whitespace from all lines.
// Handles trailing newlines from tree-sitter, expands tabs to spaces (4-char width),
// and uses rune-aware slicing for UTF-8 safety. Preserves relative indentation.
func NormalizeIndentedText(text string) string {
	text = strings.TrimSuffix(text, "\n")

	if text == "" {
		return ""
	}

	const tabWidth = 4
	text = expandTabs(text, tabWidth)

	lines := strings.Split(text, "\n")

	minIndent := -1
	for _, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		runes := []rune(line)
		indent := 0
		for _, ch := range runes {
			if ch == ' ' {
				indent++
			} else {
				break
			}
		}

		if minIndent == -1 || indent < minIndent {
			minIndent = indent
		}
	}

	if minIndent == -1 {
		return ""
	}

	result := make([]string, len(lines))
	for i, line := range lines {
		if len(strings.TrimSpace(line)) == 0 {
			result[i] = ""
		} else {
			runes := []rune(line)
			if len(runes) >= minIndent {
				result[i] = string(runes[minIndent:])
			} else {
				result[i] = line
			}
		}
	}

	normalized := strings.Join(result, "\n")
	return strings.TrimRight(normalized, " \t\n")
}

// expandTabs expands tabs to spaces, aligning to next tab stop (standard terminal behavior).
func expandTabs(text string, tabWidth int) string {
	var result strings.Builder
	col := 0

	for _, ch := range text {
		if ch == '\t' {
			// Insert spaces until next tab stop
			spaces := tabWidth - (col % tabWidth)
			for i := 0; i < spaces; i++ {
				result.WriteRune(' ')
			}
			col += spaces
		} else if ch == '\n' {
			result.WriteRune(ch)
			col = 0 // Reset column on newline
		} else {
			result.WriteRune(ch)
			col++
		}
	}

	return result.String()
}
