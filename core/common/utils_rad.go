package com

func Truncate(str string, maxLen int64) string {
	runes := []rune(str)
	if TerminalIsUtf8 {
		return string(runes[:maxLen-1]) + "…"
	} else {
		return string(runes[:maxLen-3]) + "..."
	}
}

func Reverse(str string) string {
	runeString := []rune(str)
	var reverseString string
	for i := len(runeString) - 1; i >= 0; i-- {
		reverseString += string(runeString[i])
	}
	return reverseString
}
