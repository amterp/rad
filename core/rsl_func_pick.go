package core

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

const (
	PICK_PROMPT = "prompt"
)

func runPick(i *MainInterpreter, function Token, args []interface{}, namedArgs map[string]interface{}) RslString {
	numArgs := len(args)
	if numArgs < 1 {
		i.error(function, PICK+"() takes at least one argument")
	}

	validateExpectedNamedArgs(i, function, []string{PICK_PROMPT}, namedArgs)
	parsedArgs := parsePickArgs(i, function, namedArgs)

	filters := make([]string, 0)
	switch numArgs {
	case 1:
		// no filters, leave it empty
	case 2:
		filter := args[1]
		switch filter.(type) {
		case RslString, int64, float64, bool:
			filters = append(filters, ToPrintable(filter))
		case []interface{}:
			strings, ok := AsStringArray(filter.([]interface{}))
			if !ok {
				i.error(function, PICK+"() does not allow non-string arrays as filters")
			}
			filters = strings
		default:
			i.error(function, PICK+"() does not allow non-string arrays as filters")
		}
	default:
		i.error(function, fmt.Sprintf(PICK+"() takes at most two arguments, got %v", numArgs))
	}

	switch options := args[0].(type) {
	case []interface{}:
		array, ok := AsStringArray(options)
		if !ok {
			i.error(function, PICK+fmt.Sprintf("() does not allow non-string arrays as options, got %s", TypeAsString(options)))
		}
		return pickString(i, function, parsedArgs.prompt, filters, array)
	default:
		i.error(function, PICK+"() takes a string array as the first argument")
		panic(UNREACHABLE)
	}
}

func parsePickArgs(i *MainInterpreter, function Token, args map[string]interface{}) PickNamedArgs {
	parsedArgs := PickNamedArgs{
		prompt: "Pick an option",
	}

	if prompt, ok := args[PICK_PROMPT]; ok {
		if rslString, ok := prompt.(RslString); ok {
			s := rslString.Plain()
			if StrLen(s) == 0 {
				// huh has a bug where an empty prompt cuts off an option, and it doesn't display user-typed filter
				// setting this to a space tricks huh into thinking there's a title, avoiding this issue (granted it
				// looks a bit weird but hey, the user has decided no title, what do they expect?)
				parsedArgs.prompt = " "
			} else {
				parsedArgs.prompt = s
			}
		} else {
			i.error(function, function.GetLexeme()+fmt.Sprintf("() %s must be a string, got %s", PICK_PROMPT, TypeAsString(prompt)))
		}
	}

	return parsedArgs
}

type PickNamedArgs struct {
	prompt string
}

func pickString(i *MainInterpreter, function Token, prompt string, filters []string, options []string) RslString {
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
		return NewRslString(filteredOptions[0].Value)
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

	return NewRslString(result)
}
