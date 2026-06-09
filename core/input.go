package core

import (
	"errors"
	"strings"

	com "github.com/amterp/rad/core/common"

	"github.com/amterp/radish"
)

// todo allow controlling 'yes' response?
func InputConfirm(title string, prompt string) (bool, error) {
	model := radish.NewInput().Prompt(prompt).Width(GetTermWidth())
	if !com.IsBlank(title) {
		model.Title(title)
	}
	response, err := runInput(model)
	if err != nil {
		return false, err
	}
	return response == "" || strings.HasPrefix(strings.ToLower(response), "y"), nil
}

func InputText(prompt, hint, default_ string, secret bool) (RadString, error) {
	if secret {
		return inputSecret(prompt, default_)
	}
	return inputText(prompt, hint, default_)
}

func inputSecret(prompt string, default_ string) (RadString, error) {
	// EchoNone matches the terminal convention for secrets (sudo/ssh): nothing is
	// rendered as the user types, so the value never appears in a frame.
	model := radish.NewInput().Prompt(prompt).Echo(radish.EchoNone).Width(GetTermWidth())
	return runInputDefault(model, default_)
}

func inputText(prompt string, hint string, default_ string) (RadString, error) {
	model := radish.NewInput().Prompt(prompt).Width(GetTermWidth())
	if !com.IsBlank(hint) {
		model.Placeholder(hint)
	} else if !com.IsBlank(default_) {
		model.Placeholder("Default: " + default_)
	}
	return runInputDefault(model, default_)
}

// runInputDefault runs the model and substitutes default_ when the response is empty.
func runInputDefault(model *radish.InputModel, default_ string) (RadString, error) {
	response, err := runInput(model)
	if err != nil {
		return NewRadString(""), err
	}
	if len(response) == 0 {
		response = default_
	}
	return NewRadString(response), nil
}

// runInput drives an input model through the injected interactive driver (the real
// terminal in production, a scripted driver in tests) and maps radish's outcomes to
// rad errors. A canceled prompt (Esc/Ctrl-C) surfaces as an error, matching how huh
// reported user abort.
func runInput(model *radish.InputModel) (string, error) {
	res, _, err := RInteractive.Run(model)
	if err != nil {
		if errors.Is(err, radish.ErrNotInteractive) {
			return "", errors.New("input requires an interactive terminal")
		}
		return "", err
	}
	if res.Canceled {
		return "", errors.New("input canceled")
	}
	response, _ := model.Value()
	return response, nil
}
