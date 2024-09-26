package core

import (
	"fmt"
	"github.com/charmbracelet/huh"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/lo"
)

func runPickKv(i *MainInterpreter, function Token, args []interface{}) interface{} {
	numArgs := len(args)
	if numArgs < 2 {
		i.error(function, "pick_kv() takes at least two arguments")
	}

	if numArgs > 3 {
		i.error(function, fmt.Sprintf("pick_kv() takes at most three arguments, got %v", numArgs))
	}

	filters := make([]string, 0)
	switch numArgs {
	case 2:
		// no filters, leave it empty
	case 3:
		filter := args[2]
		switch filter.(type) {
		case string, int64, float64, bool:
			filters = append(filters, ToPrintable(filter))
		case []string:
			filters = filter.([]string)
		default:
			i.error(function, "pick_kv() does not allow non-string arrays as filters")
		}
	}

	keys, ok := args[0].([]string)
	if !ok {
		i.error(function, "pick_kv() takes a string array as the first argument")
		panic(UNREACHABLE)
	}

	if len(keys) == 0 {
		i.error(function, "pick_kv() requires keys and values to have at least one element")
	}

	switch values := args[1].(type) {
	case []string:
		return pickKv(i, function, "", filters, keys, values)
	case []int64:
		return pickKv(i, function, "", filters, keys, values)
	case []float64:
		return pickKv(i, function, "", filters, keys, values)
	default:
		i.error(function, "pick_kv() takes an array as the second argument")
		panic(UNREACHABLE)
	}
}

func pickKv[T comparable](i *MainInterpreter, function Token, prompt string, filters []string, keys []string, values []T) T {
	if len(keys) != len(values) {
		i.error(function, fmt.Sprintf("pick_kv() requires keys and values to be the same length, got %d keys and %d values", len(keys), len(values)))
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
		i.error(function, fmt.Sprintf("Error running pick_kv: %v", err))
	}

	return result
}
