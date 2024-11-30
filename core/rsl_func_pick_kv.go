package core

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/lo"
)

func runPickKv(i *MainInterpreter, function Token, args []interface{}, namedArgs map[string]interface{}) interface{} {
	numArgs := len(args)
	if numArgs < 2 {
		i.error(function, PICK_KV+"() takes at least two arguments")
	}

	if numArgs > 3 {
		i.error(function, fmt.Sprintf("%s() takes at most three arguments, got %v", PICK_KV, numArgs))
	}

	validateExpectedNamedArgs(i, function, []string{PICK_PROMPT}, namedArgs)
	parsedArgs := parsePickArgs(i, function, namedArgs)

	filters := make([]string, 0)
	switch numArgs {
	case 2:
		// no filters, leave it empty
	case 3:
		filter := args[2]
		switch coerced := filter.(type) {
		case RslString:
			filters = append(filters, coerced.Plain())
		case int64, float64, bool:
			filters = append(filters, ToPrintable(coerced))
		case []interface{}:
			strings, ok := AsStringArray(coerced)
			if !ok {
				i.error(function, PICK_KV+"() does not allow non-string arrays as filters")
			}
			filters = strings
		default:
			i.error(function, PICK_KV+"() does not allow non-string arrays as filters")
		}
	}

	var keys []string
	keys, ok := args[0].([]string)
	if !ok {
		if keys, ok = AsStringArray(args[0].([]interface{})); !ok {
			i.error(function, PICK_KV+"() takes a string array as the first argument")
			panic(UNREACHABLE)
		}
	}

	if len(keys) == 0 {
		i.error(function, PICK_KV+"() requires keys and values to have at least one element")
	}

	switch values := args[1].(type) {
	case []interface{}:
		return pickKv(i, function, parsedArgs.prompt, filters, keys, values)
	default:
		i.error(function, PICK_KV+"() takes an array as the second argument")
		panic(UNREACHABLE)
	}
}

func pickKv[T comparable](i *MainInterpreter, function Token, prompt string, filters []string, keys []string, values []T) T {
	if len(keys) != len(values) {
		i.error(function, fmt.Sprintf("%s() requires keys and values to be the same length, got %d keys and %d values",
			PICK_KV, len(keys), len(values)))
	}

	filteredKeyValues := make(map[string]T)
	for index, key := range keys {
		failedAFilter := false
		for _, filter := range filters {
			if !fuzzy.MatchFold(filter, key) {
				failedAFilter = true
				break
			}
		}
		if !failedAFilter {
			filteredKeyValues[key] = values[index]
		}
	}

	if len(filteredKeyValues) == 0 {
		i.error(function, fmt.Sprintf("Filtered %d keys to 0 with filters: %q", len(keys), filters))
	}

	if len(filteredKeyValues) == 1 {
		for _, value := range filteredKeyValues {
			return value
		}
	}

	var result T
	options := lo.MapToSlice(filteredKeyValues, func(k string, v T) huh.Option[T] { return huh.NewOption(k, v) })
	err := huh.NewSelect[T]().
		Title(prompt).
		Options(options...).
		Value(&result).
		Run()

	if err != nil {
		i.error(function, fmt.Sprintf("Error running %s: %v", PICK_KV, err))
	}

	return result
}
