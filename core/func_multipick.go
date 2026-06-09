package core

import (
	"errors"
	"fmt"

	"github.com/amterp/radish"
)

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
				return f.Return(NewErrorStrf("multipick requires an interactive terminal"))
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
