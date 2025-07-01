package core

import (
	com "rad/core/common"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/samber/lo"
)

var FuncPick = BuiltInFunc{
	Name: FUNC_PICK,
	Execute: func(f FuncInvocation) RadValue {
		options := f.GetList("_options").AsStringList(false)
		filterArg := f.GetArg("_filter")

		filters := make([]string, 0)
		if !filterArg.IsNull() {
			switch coerced := filterArg.Val.(type) {
			case RadString:
				filters = append(filters, coerced.Plain())
			case *RadList:
				for _, item := range coerced.Values {
					filters = append(filters, item.RequireStr(f.i, f.callNode).Plain())
				}
			default:
				bugIncorrectTypes(FUNC_PICK)
			}
		}

		keyGroups := lo.Map(options, func(key string, _ int) []string { return []string{key} })
		str, err := pickKv(f, keyGroups, keyGroups, filters)
		if err != nil {
			return f.Return(err)
		}

		return f.Return(str[0])
	},
}

var FuncPickKv = BuiltInFunc{
	Name: FUNC_PICK_KV,
	Execute: func(f FuncInvocation) RadValue {
		keys := f.GetList("keys").AsStringList(false)
		values := f.GetList("values").Values
		filter := f.GetArg("_filter")

		filters := make([]string, 0)
		if !filter.IsNull() {
			switch coerced := filter.Val.(type) {
			case RadString:
				filters = append(filters, coerced.Plain())
			case *RadList:
				for _, item := range coerced.Values {
					filters = append(filters, item.RequireStr(f.i, f.callNode).Plain())
				}
			default:
				bugIncorrectTypes(FUNC_PICK_KV)
			}
		}

		keyGroups := lo.Map(keys, func(key string, _ int) []string { return []string{key} })
		valueGroups := lo.Map(values, func(value RadValue, _ int) []RadValue { return []RadValue{value} })

		out, err := pickKv(f, keyGroups, valueGroups, filters)
		if err != nil {
			return f.Return(err)
		}
		return f.Return(out[0])
	},
}

var FuncPickFromResource = BuiltInFunc{
	Name: FUNC_PICK_FROM_RESOURCE,
	Execute: func(f FuncInvocation) RadValue {
		path := f.GetStr("path").Plain()
		filter := f.GetArg("_filter")

		resource, err := LoadPickResource(f.i, f.callNode, path)
		if err != nil {
			return f.Return(err)
		}

		var keyGroups [][]string
		var valueGroups [][]RadValue
		for _, opt := range resource.Opts {
			keyGroups = append(keyGroups, opt.Keys)
			valueGroups = append(valueGroups, opt.Values)
		}

		filters := make([]string, 0)
		if !filter.IsNull() {
			switch coerced := filter.Val.(type) {
			case RadString:
				filters = append(filters, coerced.Plain())
			case *RadList:
				for _, item := range coerced.Values {
					filters = append(filters, item.RequireStr(f.i, f.callNode).Plain())
				}
			default:
				bugIncorrectTypes(FUNC_PICK_KV)
			}
		}

		out, err := pickKv(f, keyGroups, valueGroups, filters)

		if err != nil {
			return f.Return(err)
		}

		if len(out) == 1 {
			return newRadValues(f.i, f.callNode, out[0])
		} else {
			return newRadValues(f.i, f.callNode, out)
		}
	},
}

func pickKv[T comparable](
	f FuncInvocation,
	keyGroups [][]string,
	valueGroups [][]T,
	filters []string,
) ([]T, *RadError) {
	if len(keyGroups) != len(valueGroups) {
		return []T{}, NewErrorStrf("Number of keys and values must match, but got %s and %s",
			com.Pluralize(len(keyGroups), "key"), com.Pluralize(len(valueGroups), "value"))
	}

	prompt := f.GetStr("prompt").Plain()
	if prompt == "" {
		// huh has a bug where an empty prompt cuts off an option, and it doesn't display user-typed filter
		// setting this to a space tricks huh into thinking there's a title, avoiding this issue (granted it
		// looks a bit weird but hey, the user has decided no title, what do they expect?)
		prompt = " "
	}

	// matched values by label, plus an ordered list of labels
	matchedKeyValues := make(map[string][]T)
	orderedKeys := make([]string, 0, len(keyGroups))

	for i, keyGroup := range keyGroups {
		values := valueGroups[i]
		label := strings.Join(keyGroup, " ")

		// decide whether this one passes filters
		keep := len(filters) == 0
		if !keep {
			keep = true
			for _, filter := range filters {
				found := false
				for _, key := range keyGroup {
					if fuzzy.MatchFold(filter, key) {
						found = true
						break
					}
				}
				if !found {
					keep = false
					break
				}
			}
		}

		if keep {
			matchedKeyValues[label] = values
			orderedKeys = append(orderedKeys, label)
		}
	}

	if len(orderedKeys) == 0 {
		return []T{}, NewErrorStrf(
			"Filtered %s to 0 with filters: %v",
			com.Pluralize(len(keyGroups), "option"),
			filters,
		)
	}

	// single match? return immediately
	if len(orderedKeys) == 1 {
		return matchedKeyValues[orderedKeys[0]], nil
	}

	// build options in original order
	var selected string
	opts := make([]huh.Option[string], len(orderedKeys))
	for i, lbl := range orderedKeys {
		opts[i] = huh.NewOption(lbl, lbl)
	}

	if err := huh.NewSelect[string]().
		Title(prompt).
		Options(opts...).
		Value(&selected).
		Run(); err != nil {
		return []T{}, NewErrorStrf("Error running pick: %v", err)
	}

	return matchedKeyValues[selected], nil
}
