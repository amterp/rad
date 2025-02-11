package core

import (
	com "rad/core/common"
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

func InteractiveConfirm(title string, prompt string) (bool, error) {
	var response string
	input := huh.NewInput().
		Prompt(prompt).
		Value(&response)
	if !com.IsBlank(title) {
		input.Title(title)
	}
	err := input.Run()
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
