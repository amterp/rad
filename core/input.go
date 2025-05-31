package core

import (
	com "rad/core/common"
	"strings"

	"github.com/charmbracelet/huh"
)

// todo allow controlling 'yes' response?
func InputConfirm(title string, prompt string) (bool, error) {
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

func InputText(prompt, hint, default_ string) (RadString, error) {
	var response string
	input := huh.NewInput().
		Prompt(prompt).
		Value(&response)

	if com.IsBlank(hint) {
		input.Placeholder("Default: " + default_)
	} else {
		input.Placeholder(hint)
	}

	err := input.Run()
	if len(response) == 0 {
		response = default_
	}
	return NewRadString(response), err
}
