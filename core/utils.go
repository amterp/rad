package core

import com "rad/core/common"

func ToStringArrayQuoteStr[T any](v []T, quoteStrings bool) []string {
	output := make([]string, len(v))
	for i, val := range v {
		output[i] = ToPrintableQuoteStr(val, quoteStrings)
	}
	return output
}

func Truncate(str string, maxLen int64) string {
	if com.TerminalIsUtf8 {
		str = str[:maxLen-1]
		str += "…"
	} else {
		str = str[:maxLen-3]
		str += "..."
	}
	return str
}
