package core

import (
	"github.com/charmbracelet/huh"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncPick = Func{
	Name:             FUNC_PICK,
	ReturnValues:     ONE_RETURN_VAL,
	RequiredArgCount: 1,
	ArgTypes:         [][]RslTypeEnum{{RslListT}, {RslStringT, RslListT}},
	NamedArgs: map[string][]RslTypeEnum{
		namedArgPrompt: {RslStringT},
	},
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, namedArgs map[string]namedArg) []RslValue {
		optionsArg := args[0]

		filters := make([]string, 0)
		if len(args) == 2 {
			filteringNode := args[1]

			switch coerced := filteringNode.value.Val.(type) {
			case RslString:
				filters = append(filters, coerced.Plain())
			case *RslList:
				for _, item := range coerced.Values {
					if str, ok := item.Val.(RslString); ok {
						filters = append(filters, str.Plain())
					} else {
						i.errorf(filteringNode.node,
							"All filters must be strings, but got %q: %v", TypeAsString(item), item)
					}
				}
			default:
				bugIncorrectTypes(FUNC_PICK)
			}
		}

		options := optionsArg.value.RequireList(i, optionsArg.node).AsStringList(false)
		str := pickKv(i, callNode, options, options, filters, namedArgs)
		return newRslValues(i, callNode, str)
	},
}

var FuncPickKv = Func{
	Name:             FUNC_PICK_KV,
	ReturnValues:     ONE_RETURN_VAL,
	RequiredArgCount: 2,
	ArgTypes:         [][]RslTypeEnum{{RslListT}, {RslListT}, {RslStringT, RslListT}},
	NamedArgs: map[string][]RslTypeEnum{
		namedArgPrompt: {RslStringT},
	},
	Execute: func(i *Interpreter, callNode *ts.Node, args []positionalArg, namedArgs map[string]namedArg) []RslValue {
		keyArgs := args[0]
		valueArgs := args[1]

		filters := make([]string, 0)
		if len(args) == 3 {
			filteringNode := args[2]

			switch coerced := filteringNode.value.Val.(type) {
			case RslString:
				filters = append(filters, coerced.Plain())
			case *RslList:
				for _, item := range coerced.Values {
					if str, ok := item.Val.(RslString); ok {
						filters = append(filters, str.Plain())
					} else {
						i.errorf(filteringNode.node,
							"All filters must be strings, but got %q: %v", TypeAsString(item), item)
					}
				}
			default:
				bugIncorrectTypes(FUNC_PICK_KV)
			}
		}

		keys := keyArgs.value.RequireList(i, keyArgs.node).AsStringList(false)
		values := valueArgs.value.RequireList(i, valueArgs.node).Values
		value := pickKv(i, callNode, keys, values, filters, namedArgs)
		return newRslValues(i, callNode, value)
	},
}

func pickKv[T comparable](
	i *Interpreter,
	callNode *ts.Node,
	keys []string,
	values []T,
	filters []string,
	namedArgs map[string]namedArg,
) T {
	prompt := "Pick an option"
	if promptArg, ok := namedArgs[namedArgPrompt]; ok {
		prompt = promptArg.value.RequireStr(i, promptArg.valueNode).Plain()
		if prompt == "" {
			// huh has a bug where an empty prompt cuts off an option, and it doesn't display user-typed filter
			// setting this to a space tricks huh into thinking there's a title, avoiding this issue (granted it
			// looks a bit weird but hey, the user has decided no title, what do they expect?)
			prompt = " "
		}
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
		// todo potentially allow recovering from this?
		i.errorf(callNode, "Filtered %s to 0 with filters: %v", Pluralize(len(keys), "option"), filters)
	}

	if len(filteredKeyValues) == 1 {
		return filteredKeyValues[lo.Keys(filteredKeyValues)[0]]
	}

	var result T
	options := lo.MapToSlice(filteredKeyValues, func(k string, v T) huh.Option[T] { return huh.NewOption(k, v) })
	err := huh.NewSelect[T]().
		Title(prompt).
		Options(options...).
		Value(&result).
		Run()

	if err != nil {
		// todo If user aborts, this gets triggered (probably should just 'silently' exit if user aborts)
		i.errorf(callNode, "Error running pick: %v", err)
	}

	return result
}
