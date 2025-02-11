package core

import (
	"strings"

	"github.com/charmbracelet/huh"
)

func ToStringArrayQuoteStr[T any](v []T, quoteStrings bool) []string {
	output := make([]string, len(v))
	for i, val := range v {
		output[i] = ToPrintableQuoteStr(val, quoteStrings)
	}
	return output
}

func InteractiveConfirm(prompt string) (bool, error) {
	var response string
	err := huh.NewInput().
		Prompt(prompt).
		Value(&response).
		Run()
	return strings.HasPrefix(strings.ToLower(response), "y"), err
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
