package core

import (
	com "rad/core/common"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/lo"
	ts "github.com/tree-sitter/go-tree-sitter"
)

var FuncPick = BuiltInFunc{
	Name: FUNC_PICK,
	Execute: func(f FuncInvocation) RadValue {
		optionsArg := f.args[0]
		filteringArg := tryGetArg(1, f.args)

		filters := make([]string, 0)
		if filteringArg != nil {
			switch coerced := filteringArg.value.Val.(type) {
			case RadString:
				filters = append(filters, coerced.Plain())
			case *RadList:
				for _, item := range coerced.Values {
					if str, ok := item.Val.(RadString); ok {
						filters = append(filters, str.Plain())
					} else {
						f.i.errorf(filteringArg.node,
							"All filters must be strings, but got %q: %v", TypeAsString(item), item)
					}
				}
			default:
				bugIncorrectTypes(FUNC_PICK)
			}
		}

		keys := optionsArg.value.RequireList(f.i, optionsArg.node).AsStringList(false)
		keyGroups := lo.Map(keys, func(key string, _ int) []string { return []string{key} })
		str := pickKv(f.i, f.callNode, keyGroups, keyGroups, filters, f.namedArgs)[0]
		return newRadValues(f.i, f.callNode, str)
	},
}

var FuncPickKv = BuiltInFunc{
	Name: FUNC_PICK_KV,
	Execute: func(f FuncInvocation) RadValue {
		keyArgs := f.args[0]
		valueArgs := f.args[1]
		filteringArg := tryGetArg(2, f.args)

		filters := make([]string, 0)
		if filteringArg != nil {
			switch coerced := filteringArg.value.Val.(type) {
			case RadString:
				filters = append(filters, coerced.Plain())
			case *RadList:
				for _, item := range coerced.Values {
					if str, ok := item.Val.(RadString); ok {
						filters = append(filters, str.Plain())
					} else {
						f.i.errorf(filteringArg.node,
							"All filters must be strings, but got %q: %v", TypeAsString(item), item)
					}
				}
			default:
				bugIncorrectTypes(FUNC_PICK_KV)
			}
		}

		keys := keyArgs.value.RequireList(f.i, keyArgs.node).AsStringList(false)
		values := valueArgs.value.RequireList(f.i, valueArgs.node).Values

		keyGroups := lo.Map(keys, func(key string, _ int) []string { return []string{key} })
		valueGroups := lo.Map(values, func(value RadValue, _ int) []RadValue { return []RadValue{value} })

		out := pickKv(f.i, f.callNode, keyGroups, valueGroups, filters, f.namedArgs)[0]
		return newRadValues(f.i, f.callNode, out)
	},
}

var FuncPickFromResource = BuiltInFunc{
	Name: FUNC_PICK_FROM_RESOURCE,
	Execute: func(f FuncInvocation) RadValue {
		fileArg := f.args[0]
		filteringArg := tryGetArg(1, f.args)

		filePath := fileArg.value.RequireStr(f.i, fileArg.node).Plain()

		resource := LoadPickResource(f.i, f.callNode, filePath)
		var keyGroups [][]string
		var valueGroups [][]RadValue
		for _, opt := range resource.Opts {
			keyGroups = append(keyGroups, opt.Keys)
			valueGroups = append(valueGroups, opt.Values)
		}

		filters := make([]string, 0)
		if filteringArg != nil {
			switch coerced := filteringArg.value.Val.(type) {
			case RadString:
				filters = append(filters, coerced.Plain())
			case *RadList:
				for _, item := range coerced.Values {
					if str, ok := item.Val.(RadString); ok {
						filters = append(filters, str.Plain())
					} else {
						f.i.errorf(filteringArg.node,
							"All filters must be strings, but got %q: %v", TypeAsString(item), item)
					}
				}
			default:
				bugIncorrectTypes(FUNC_PICK_KV)
			}
		}

		out := pickKv(f.i, f.callNode, keyGroups, valueGroups, filters, f.namedArgs)
		if len(out) == 1 {
			return newRadValues(f.i, f.callNode, out[0])
		} else {
			return newRadValues(f.i, f.callNode, out)
		}
	},
}

func pickKv[T comparable](
	i *Interpreter,
	callNode *ts.Node,
	keyGroups [][]string,
	valueGroups [][]T,
	filters []string,
	namedArgs map[string]namedArg,
) []T {
	if len(keyGroups) != len(valueGroups) {
		i.errorf(callNode, "Number of keys and values must match, but got %s and %s",
			com.Pluralize(len(keyGroups), "key"), com.Pluralize(len(valueGroups), "value"))
	}

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

	matchedKeyValues := make(map[string][]T)
	for index, keyGroup := range keyGroups {
		values := valueGroups[index]
		entryKey := strings.Join(lo.Map(values, func(v T, _ int) string { return ToPrintableQuoteStr(v, false) }), " ")
		entryKey = entryKey + " (" + strings.Join(keyGroup, " ") + ")"
		for _, key := range keyGroup {

			if len(filters) == 0 {
				matchedKeyValues[entryKey] = values
			} else {
				failedAFilter := false
				for _, filter := range filters {
					if !fuzzy.MatchFold(filter, key) {
						failedAFilter = true
						break
					}
				}
				if !failedAFilter {
					matchedKeyValues[entryKey] = values
				}
			}
		}
	}

	if len(matchedKeyValues) == 0 {
		// todo potentially allow recovering from this?
		i.errorf(callNode, "Filtered %s to 0 with filters: %v", com.Pluralize(len(keyGroups), "option"), filters)
	}

	if len(matchedKeyValues) == 1 {
		return matchedKeyValues[lo.Keys(matchedKeyValues)[0]]
	}

	var result string
	options := lo.Map(
		lo.Keys(matchedKeyValues),
		func(k string, _ int) huh.Option[string] { return huh.NewOption(k, k) },
	)
	err := huh.NewSelect[string]().
		Title(prompt).
		Options(options...).
		Value(&result).
		Run()

	if err != nil {
		// todo If user aborts, this gets triggered (probably should just 'silently' exit if user aborts)
		i.errorf(callNode, "Error running pick: %v", err)
	}

	return matchedKeyValues[result]
}
