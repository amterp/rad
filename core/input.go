package core

import (
	"fmt"
	"os"
	com "rad/core/common"
	"strings"

	"github.com/charmbracelet/huh"
	"golang.org/x/term"
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

func InputText(prompt, hint, default_ string, secret bool) (RadString, error) {
	if secret {
		return inputSecret(prompt, default_)
	}

	return inputText(prompt, hint, default_)
}

// todo colors don't match huh non-secret version
func inputSecret(prompt string, default_ string) (RadString, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Println()
	if err != nil {
		return NewRadString(""), err
	}
	response := string(bytePassword)
	if len(response) == 0 {
		response = default_
	}
	return NewRadString(response), nil
}

func inputText(prompt string, hint string, default_ string) (RadString, error) {
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
