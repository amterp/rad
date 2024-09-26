package core

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

func runPick(i *MainInterpreter, function Token, values []interface{}) interface{} {
	numArgs := len(values)
	if numArgs < 1 {
		i.error(function, "pick() takes at least one argument")
	}

	filters := make([]string, 0)
	switch numArgs {
	case 1:
		// no filters, leave it empty
	case 2:
		filter := values[1]
		switch filter.(type) {
		case string, int64, float64, bool:
			filters = append(filters, ToPrintable(filter))
		case []string:
			filters = filter.([]string)
		default:
			i.error(function, "pick() does not allow non-string arrays as filters")
		}
	default:
		i.error(function, fmt.Sprintf("pick() takes at most two arguments, got %v", numArgs))
	}

	switch options := values[0].(type) {
	case []string:
		// todo prompt should be a named/optional arg when we support that, i.e. pick(options, prompt="foo")
		return pickString(i, function, "", filters, options)
	default:
		i.error(function, "pick() takes a string array as the first argument")
		panic(UNREACHABLE)
	}
}

func pickString(i *MainInterpreter, function Token, prompt string, filters []string, options []string) string {
	var filteredOptions []huh.Option[string]
	for _, option := range options {
		failedAFilter := false
		for _, filter := range filters {
			if !fuzzy.MatchFold(filter, option) {
				failedAFilter = true
				break
			}
		}
		if !failedAFilter {
			filteredOptions = append(filteredOptions, huh.NewOption(option, option))
		}
	}

	if len(filteredOptions) == 0 {
		i.error(function, fmt.Sprintf("Filtered %d options to 0 with filters: %v", len(options), filters))
	}

	if len(filteredOptions) == 1 {
		return filteredOptions[0].Value
	}

	var result string
	// todo this probably needs to be mocked out for testing, i don't see a built-in way with huh to
	//  e.g. provide input as part of a unit test (particularly when using stdin for the RSL script)
	err := huh.NewSelect[string]().
		Title(prompt).
		Options(filteredOptions...).
		Value(&result).
		Run()

	if err != nil {
		i.error(function, fmt.Sprintf("Error running pick: %v", err))
	}

	return result
}
