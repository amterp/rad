package core

func ToStringArrayQuoteStr[T any](v []T, quoteStrings bool) []string {
	output := make([]string, len(v))
	for i, val := range v {
		output[i] = ToPrintableQuoteStr(val, quoteStrings)
	}
	return output
}

func Truncate(str string, maxLen int64) string {
	if terminalSupportsUtf8 {
		str = str[:maxLen-1]
		str += "…"
	} else {
		str = str[:maxLen-3]
		str += "..."
	}
	return str
}
