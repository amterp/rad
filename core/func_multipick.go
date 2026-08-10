package core

import (
	"errors"
	"fmt"
	"strings"

	com "github.com/amterp/rad/core/common"
	"github.com/amterp/rad/rts/prompts"
	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/radish"
)

// multipickFromReply applies a command-line selection, enforcing the same
// bounds the interactive model would.
//
// radish blocks a bounds violation silently - Max refuses the toggle, Min
// refuses to submit - which is right at a keyboard, where the user simply sees
// nothing happen and tries again. A supplied answer has nobody to notice, so
// here it has to be an error rather than a quietly truncated selection.
func multipickFromReply(
	selected, options []string,
	min int64,
	max *int64,
	prompt string,
) (*RadList, *RadError) {
	seen := make(map[string]bool, len(selected))
	for _, sel := range selected {
		found := false
		for _, opt := range options {
			if opt == sel {
				found = true
				break
			}
		}
		if !found {
			return nil, NewErrorStrf(
				"%q is not one of the options for the multipick %q. Options: %s. "+
					"Answers must match exactly - rad won't guess which one you meant",
				sel, prompt, strings.Join(options, ", "),
			).SetCode(rl.ErrPromptUnanswerable)
		}
		// The model toggles, so it can't return an option twice. Left alone, a
		// repeat would also satisfy `min` with fewer distinct choices than asked.
		if seen[sel] {
			return nil, NewErrorStrf(
				"%q is selected twice for the multipick %q; list each option at most once",
				sel, prompt,
			).SetCode(rl.ErrPromptUnanswerable)
		}
		seen[sel] = true
	}

	if int64(len(selected)) < min {
		return nil, NewErrorStrf(
			"the multipick %q needs at least %s, but the answer gave %d",
			prompt, com.Pluralize(int(min), "selection"), len(selected),
		).SetCode(rl.ErrPromptUnanswerable)
	}
	if max != nil && int64(len(selected)) > *max {
		return nil, NewErrorStrf(
			"the multipick %q allows at most %s, but the answer gave %d",
			prompt, com.Pluralize(int(*max), "selection"), len(selected),
		).SetCode(rl.ErrPromptUnanswerable)
	}

	result := NewRadList()
	for _, item := range selected {
		result.Append(newRadValueStr(item))
	}
	return result, nil
}

var FuncMultipick = BuiltInFunc{
	Name: FUNC_MULTIPICK,
	Execute: func(f FuncInvocation) RadValue {
		// Extract options
		options := f.GetList("_options").AsStringList(false)

		// Validate options not empty
		if len(options) == 0 {
			return f.Return(NewErrorStrf("Cannot multipick from empty options list"))
		}

		// Extract parameters
		minArg := f.GetArg("min")
		maxArg := f.GetArg("max")
		promptArg := f.GetArg("prompt")

		// Get min value (default 0)
		min := int64(0)
		if !minArg.IsNull() {
			min = minArg.RequireInt(f.i, f.callNode)
		}

		// Validate min
		if min < 0 {
			return f.Return(NewErrorStrf("min must be non-negative, got %d", min))
		}

		// Get max value (optional)
		var max *int64
		if !maxArg.IsNull() {
			maxVal := maxArg.RequireInt(f.i, f.callNode)
			max = &maxVal

			// Validate max
			if maxVal <= 0 {
				return f.Return(NewErrorStrf("max must be positive, got %d", maxVal))
			}

			// Validate min <= max
			if min > maxVal {
				return f.Return(NewErrorStrf("min (%d) cannot be greater than max (%d)", min, maxVal))
			}
		}

		// Validate min against number of options
		if min > int64(len(options)) {
			if min == 1 {
				return f.Return(NewErrorStrf("min is 1 but there are no options available"))
			} else {
				return f.Return(NewErrorStrf("min is %d but only %d options available", min, len(options)))
			}
		}

		// Generate smart default prompt if not provided
		var prompt string
		if promptArg.IsNull() {
			prompt = generateMultipickPrompt(min, max)
		} else {
			prompt = f.GetStr("prompt").Plain()
		}

		if answer, outcome := takeReply(f.callNode); outcome != prompts.NoReply {
			if outcome != prompts.Answered {
				return f.Return(unansweredPromptErr(outcome, fmt.Sprintf("The multipick %q", prompt)))
			}
			selected, radErr := multipickFromReply(answer.List, options, min, max, prompt)
			if radErr != nil {
				return f.Return(radErr)
			}
			return f.Return(selected)
		}

		// radish enforces the bounds directly: Max blocks toggling past the limit,
		// Min gates submit until satisfied. No post-submit validation is needed - the
		// returned selection is always within [min, max].
		model := radish.NewMultiSelect().
			Title(prompt).
			Options(options...).
			Min(int(min)).
			Width(GetTermWidth())
		if max != nil {
			model.Max(int(*max))
		}

		res, _, err := RInteractive.Run(model)
		if err != nil {
			if errors.Is(err, radish.ErrNotInteractive) {
				return f.Return(unansweredPromptErr(prompts.NoReply, fmt.Sprintf("The multipick %q", prompt)))
			}
			return f.Return(NewErrorStrf("Error running multipick: %v", err))
		}
		if res.Canceled {
			return f.Return(NewErrorStrf("multipick canceled"))
		}

		// Convert to RadList
		result := NewRadList()
		for _, item := range model.Selected() {
			result.Append(newRadValueStr(item))
		}

		return f.Return(result)
	},
}

// generateMultipickPrompt creates a smart default prompt based on min/max constraints
func generateMultipickPrompt(min int64, max *int64) string {
	if max == nil {
		// No max limit
		if min == 0 {
			return "Select options"
		} else if min == 1 {
			return "Select at least 1 option"
		} else {
			return fmt.Sprintf("Select at least %d options", min)
		}
	} else {
		// Has max limit
		if min == *max {
			// Exactly N selections required
			if min == 1 {
				return "Select 1 option"
			} else {
				return fmt.Sprintf("Select %d options", min)
			}
		} else if min == 0 {
			if *max == 1 {
				return "Select up to 1 option"
			} else {
				return fmt.Sprintf("Select up to %d options", *max)
			}
		} else {
			// Range of selections
			return fmt.Sprintf("Select %d-%d options", min, *max)
		}
	}
}
