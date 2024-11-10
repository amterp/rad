package core

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"strings"
)

func runConfirm(i *MainInterpreter, function Token, args []interface{}) bool {
	if len(args) > 1 {
		i.error(function, CONFIRM+fmt.Sprintf("() takes at most 1 arg, got %d", len(args)))
	}

	prompt := "Confirm? [y/n] "

	if len(args) == 1 {
		prompt = ToPrintable(args[0])
	}

	var response string
	err := huh.NewInput().
		Prompt(prompt).
		Value(&response).
		Run()
	if err != nil {
		i.error(function, fmt.Sprintf("Error reading input: %v", err))
	}

	return strings.HasPrefix(strings.ToLower(response), "y")
}
