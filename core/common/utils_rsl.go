package com

func Truncate(str string, maxLen int64) string {
	if TerminalIsUtf8 {
		str = str[:maxLen-1]
		str += "…"
	} else {
		str = str[:maxLen-3]
		str += "..."
	}
	return str
}

func Reverse(str string) string {
	runeString := []rune(str)
	var reverseString string
	for i := len(runeString) - 1; i >= 0; i-- {
		reverseString += string(runeString[i])
	}
	return reverseString
}
