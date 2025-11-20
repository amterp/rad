package core

import (
	"fmt"

	"github.com/charmbracelet/huh"
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
			if prompt == "" {
				// huh has a bug where an empty prompt cuts off an option
				prompt = " "
			}
		}

		// Build options for huh
		opts := make([]huh.Option[string], len(options))
		for i, opt := range options {
			opts[i] = huh.NewOption(opt, opt)
		}

		// Create multi-select with validation
		multiSelect := huh.NewMultiSelect[string]().
			Title(prompt).
			Options(opts...).
			Validate(func(selected []string) error {
				count := int64(len(selected))

				// Special case: exact count required (min == max)
				if max != nil && min == *max {
					if count != min {
						if min == 1 {
							return fmt.Errorf("Must select exactly 1 option, but selected %d", count)
						} else {
							return fmt.Errorf("Must select exactly %d options, but selected %d", min, count)
						}
					}
					return nil
				}

				// Check minimum constraint
				if count < min {
					if min == 1 {
						return fmt.Errorf("Must select at least 1 option, but only selected %d", count)
					} else {
						return fmt.Errorf("Must select at least %d options, but only selected %d", min, count)
					}
				}

				// Check maximum constraint (huh's Limit handles UI, but validate for consistency)
				if max != nil && count > *max {
					if *max == 1 {
						return fmt.Errorf("Must select at most 1 option, but selected %d", count)
					} else {
						return fmt.Errorf("Must select at most %d options, but selected %d", *max, count)
					}
				}

				return nil
			})

		// Apply limit if max is set
		if max != nil {
			multiSelect = multiSelect.Limit(int(*max))
		}

		// Execute the selection
		var selected []string
		multiSelect = multiSelect.Value(&selected)

		if err := multiSelect.Run(); err != nil {
			return f.Return(NewErrorStrf("Error running multipick: %v", err))
		}

		// Convert to RadList
		result := NewRadList()
		for _, item := range selected {
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
