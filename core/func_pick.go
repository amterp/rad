package core

import (
	"errors"
	"fmt"
	"strings"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/prompts"
	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/radish"
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
		prioExact := f.GetBool(namedArgPreferExact)
		str, err := pickKv(f, keyGroups, keyGroups, filters, prioExact)
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

		prioExact := f.GetBool(namedArgPreferExact)
		out, err := pickKv(f, keyGroups, valueGroups, filters, prioExact)
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

		prioExact := f.GetBool(namedArgPreferExact)
		out, err := pickKv(f, keyGroups, valueGroups, filters, prioExact)

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
	prioExact bool,
) ([]T, *RadError) {
	if len(keyGroups) != len(valueGroups) {
		return []T{}, NewErrorStrf("Number of keys and values must match, but got %s and %s",
			com.Pluralize(len(keyGroups), "key"), com.Pluralize(len(valueGroups), "value"))
	}

	prompt := f.GetStr("prompt").Plain()

	// matched values by label, plus an ordered list of labels
	matchedKeyValues := make(map[string][]T)
	orderedKeys := make([]string, 0, len(keyGroups))
	hasExactMatch := make(map[string]bool)

	for i, keyGroup := range keyGroups {
		values := valueGroups[i]
		label := strings.Join(keyGroup, " ")

		// decide whether this one passes filters
		keep := len(filters) == 0
		foundExactMatch := false

		if !keep {
			keep = true
			for _, filter := range filters {
				filterMatched := false
				for _, key := range keyGroup {
					if FuzzyMatchFold(filter, key) {
						filterMatched = true
					}
					if strings.EqualFold(filter, key) {
						foundExactMatch = true
					}
				}
				if !filterMatched {
					keep = false
					break
				}
			}
		}

		if keep {
			matchedKeyValues[label] = values
			orderedKeys = append(orderedKeys, label)
			if foundExactMatch {
				hasExactMatch[label] = true
			}
		}
	}

	if len(orderedKeys) == 0 {
		return []T{}, NewErrorStrf(
			"Filtered %s to 0 with filters: %v",
			com.Pluralize(len(keyGroups), "option"),
			filters,
		)
	}

	// A filter can settle the choice on its own, in which case nothing is asked.
	settled, settledOn := "", len(orderedKeys) == 1
	if settledOn {
		settled = orderedKeys[0]
	} else if prioExact {
		// exact match priority: if exactly one entry matches a key exactly, pick it
		var exactMatchLabels []string
		for _, lbl := range orderedKeys {
			if hasExactMatch[lbl] {
				exactMatchLabels = append(exactMatchLabels, lbl)
			}
		}
		if len(exactMatchLabels) == 1 {
			settled, settledOn = exactMatchLabels[0], true
		} else if len(exactMatchLabels) > 1 {
			orderedKeys = exactMatchLabels // narrow picker to exact matches only
		}
	}

	if settledOn {
		// The script's own filter has decided; nothing is asked, so a caller who
		// supplied no answer is fine. That is what keeps --reply-na valid here.
		//
		// An answer that *is* waiting is still consumed, which is what keeps a
		// per-line queue in step with executions - a pass that quietly took
		// nothing would hand every later answer to the wrong pass. It is only
		// dropped when it names what the filter chose. Otherwise the caller
		// asked for one option, the filter took another, and going quiet acts
		// on the filter's choice while the caller believes theirs was used.
		//
		// Agreement stays silent, so a pick whose list happens to hold one entry
		// keeps working. The same wrong answer already fails below when the pick
		// does prompt, so this closes a gap rather than tightening a rule.
		if replyPending(f.callNode) {
			if answer, outcome := takeReply(f.callNode); outcome == prompts.Answered &&
				answer.Text != settled {
				return []T{}, NewErrorStrf(
					"The pick %q settled on %q without asking, but --reply said %q. rad won't "+
						"override the script's own filter: change what the script filters on, "+
						"or answer with %q%s",
					prompt, settled, answer.Text, settled, retryRepeatsHint(),
				).SetCode(rl.ErrPromptUnanswerable)
			}
		}
		return matchedKeyValues[settled], nil
	}

	// Only now is a prompt actually going to happen.
	if answer, outcome := takeReply(f.callNode); outcome != prompts.NoReply {
		if outcome != prompts.Answered {
			return []T{}, unansweredPromptErr(f.callNode, outcome, fmt.Sprintf("The pick %q", prompt))
		}
		// Checked against the picker's own list rather than every match, so an
		// answer can't name an option prefer_exact already narrowed away.
		if !lo.Contains(orderedKeys, answer.Text) {
			// Options here were computed at runtime, so this could not have been
			// caught up front - name them now so one re-run can fix it.
			return []T{}, NewErrorStrf(
				"%q is not one of the options for the pick %q. Options: %s. "+
					"Answers must match exactly - rad won't guess which one you meant",
				answer.Text, prompt, strings.Join(orderedKeys, ", "),
			).SetCode(rl.ErrPromptUnanswerable)
		}
		return matchedKeyValues[answer.Text], nil
	}

	model := radish.NewSelect().
		Title(prompt).
		Options(orderedKeys...).
		Matcher(radishPickMatcher).
		Width(GetTermWidth())

	res, _, err := RInteractive.Run(model)
	if err != nil {
		if errors.Is(err, radish.ErrNotInteractive) {
			return []T{}, unansweredPromptErr(f.callNode, prompts.NoReply, fmt.Sprintf("The pick %q", prompt))
		}
		return []T{}, NewErrorStrf("Error running pick: %v", err)
	}
	if res.Canceled {
		return []T{}, NewErrorStrf("pick canceled")
	}

	// model is the same *SelectModel we built and Run mutated in place, so we read
	// the result directly - no type assertion on the returned Model needed.
	selected, _ := model.Selected()
	return matchedKeyValues[selected], nil
}

// radishPickMatcher governs live, type-to-filter matching inside the picker. It
// mirrors rad's startup filter semantics (case-insensitive fuzzy match) and ranks
// an exact match first so it lands under the cursor.
func radishPickMatcher(filter, label string) (bool, int) {
	if filter == "" {
		return true, 1
	}
	if !FuzzyMatchFold(filter, label) {
		return false, 0
	}
	if strings.EqualFold(filter, label) {
		return true, 0
	}
	return true, 1
}
