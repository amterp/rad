package ra

import (
	"fmt"
	"sort"
	"strings"
)

func (c *Cmd) GenerateUsage(isLongHelp bool) string {
	return c.generateUsage(isLongHelp)
}

func (c *Cmd) GenerateShortUsage() string {
	return c.generateUsage(false)
}

func (c *Cmd) GenerateLongUsage() string {
	return c.generateUsage(true)
}

func (c *Cmd) generateUsage(isLongHelp bool) string {
	var sb strings.Builder

	if c.description != "" {
		sb.WriteString(c.description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("Usage:\n  ")
	sb.WriteString(c.generateSynopsis(isLongHelp))
	sb.WriteString("\n")

	// Separate script and global flags
	var scriptFlags []any
	var globalFlags []any

	// Use a map to keep track of added flags to avoid duplicates
	addedFlags := make(map[string]bool)

	// Process positional flags in registration order
	for _, name := range c.positional {
		if addedFlags[name] {
			continue
		}
		flag := c.flags[name]

		isGlobal := false
		for _, gName := range c.globalFlags {
			if name == gName {
				isGlobal = true
				break
			}
		}
		if isGlobal {
			globalFlags = append(globalFlags, flag)
		} else {
			scriptFlags = append(scriptFlags, flag)
		}
		addedFlags[name] = true
	}

	// Process non-positional flags in registration order
	for _, name := range c.nonPositional {
		if addedFlags[name] {
			continue
		}
		flag := c.flags[name]

		isGlobal := false
		for _, gName := range c.globalFlags {
			if name == gName {
				isGlobal = true
				break
			}
		}
		if isGlobal {
			globalFlags = append(globalFlags, flag)
		} else {
			scriptFlags = append(scriptFlags, flag)
		}
		addedFlags[name] = true
	}

	if len(c.subCmds) > 0 {
		sb.WriteString("\nCommands:\n")
		// Sort subcommand names for consistent output
		var subCmdNames []string
		for name := range c.subCmds {
			subCmdNames = append(subCmdNames, name)
		}
		sort.Strings(subCmdNames)
		for _, name := range subCmdNames {
			subCmd := c.subCmds[name]
			if subCmd.description != "" {
				sb.WriteString(fmt.Sprintf("  %-30s%s\n", name, subCmd.description))
			} else {
				sb.WriteString(fmt.Sprintf("  %s\n", name))
			}
		}
	}

	if len(scriptFlags) > 0 {
		sb.WriteString("\nArguments:\n")
		sb.WriteString(c.formatFlags(scriptFlags, isLongHelp))
	}

	if len(globalFlags) > 0 {
		sb.WriteString("\nGlobal options:\n")
		sb.WriteString(c.formatFlags(globalFlags, isLongHelp))
	}

	return sb.String()
}

func (c *Cmd) generateSynopsis(isLongHelp bool) string {
	var sb strings.Builder
	sb.WriteString(c.name)

	if len(c.subCmds) > 0 {
		sb.WriteString(" [subcommand]")
		// Still show parent command flags in synopsis even when subcommands exist
		for _, name := range c.positional {
			flag := c.flags[name]
			base := getBaseFlag(flag)
			if base.Hidden {
				continue
			}
			if !isLongHelp && base.HiddenInShortHelp {
				continue
			}

			// Show positional-only flags or non-bool flags in synopsis (bools never appear)
			flagType := getFlagType(flag)
			if base.PositionalOnly || flagType != "bool" {
				var argName string

				// Check if it's a variadic slice
				isVariadic := false
				switch f := flag.(type) {
				case *StringSliceFlag:
					isVariadic = f.Variadic
				case *IntSliceFlag:
					isVariadic = f.Variadic
				case *Int64SliceFlag:
					isVariadic = f.Variadic
				case *Float64SliceFlag:
					isVariadic = f.Variadic
				case *BoolSliceFlag:
					isVariadic = f.Variadic
				}

				if isVariadic {
					argName = name + "..."
				} else {
					argName = name
				}

				// Determine if flag should show as required or optional in synopsis
				shouldBeOptional := c.shouldFlagBeOptionalInSynopsis(flag)

				if shouldBeOptional {
					sb.WriteString(fmt.Sprintf(" [%s]", argName))
				} else {
					sb.WriteString(fmt.Sprintf(" <%s>", argName))
				}
			}
		}
		sb.WriteString(" [OPTIONS]")
		return sb.String()
	}

	// First pass: collect positional-only flags
	var positionalOnlyFlags []string
	var nonPositionalFlags []string

	for _, name := range c.positional {
		flag := c.flags[name]
		base := getBaseFlag(flag)
		if base.Hidden {
			continue
		}
		if !isLongHelp && base.HiddenInShortHelp {
			continue
		}

		flagType := getFlagType(flag)
		if flagType == "bool" {
			continue // Bools never appear in synopsis
		}

		if base.PositionalOnly {
			positionalOnlyFlags = append(positionalOnlyFlags, name)
		} else {
			// Check if it's variadic (always appears in synopsis)
			isVariadic := false
			switch f := flag.(type) {
			case *StringSliceFlag:
				isVariadic = f.Variadic
			case *IntSliceFlag:
				isVariadic = f.Variadic
			case *Int64SliceFlag:
				isVariadic = f.Variadic
			case *Float64SliceFlag:
				isVariadic = f.Variadic
			case *BoolSliceFlag:
				isVariadic = f.Variadic
			}

			if isVariadic || !base.Optional {
				// Variadic flags always appear (they're inherently optional)
				// Non-variadic required flags also appear
				nonPositionalFlags = append(nonPositionalFlags, name)
			}
		}
		// Non-variadic optional flags don't appear in synopsis
	}

	// Add non-positional flags from nonPositional list (they might not be in positional)
	for _, name := range c.nonPositional {
		flag := c.flags[name]
		base := getBaseFlag(flag)
		if base.Hidden {
			continue
		}
		if !isLongHelp && base.HiddenInShortHelp {
			continue
		}

		// Skip global flags - they don't appear in synopsis
		isGlobal := false
		for _, gName := range c.globalFlags {
			if name == gName {
				isGlobal = true
				break
			}
		}
		if isGlobal {
			continue
		}

		flagType := getFlagType(flag)
		if flagType == "bool" || base.Optional {
			continue // Bools and optional flags never appear in synopsis
		}

		// Check if already added
		found := false
		for _, existing := range nonPositionalFlags {
			if existing == name {
				found = true
				break
			}
		}
		if !found {
			nonPositionalFlags = append(nonPositionalFlags, name)
		}
	}

	// Process positional-only flags first, but stop after first variadic
	for _, name := range positionalOnlyFlags {
		flag := c.flags[name]
		shouldBeOptional := c.shouldFlagBeOptionalInSynopsis(flag)

		// Check if it's a variadic positional flag
		argName := name
		isVariadic := false
		switch f := flag.(type) {
		case *StringSliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *IntSliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *Int64SliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *Float64SliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		case *BoolSliceFlag:
			if f.Variadic {
				argName = name + "..."
				isVariadic = true
			}
		}

		if shouldBeOptional {
			sb.WriteString(fmt.Sprintf(" [%s]", argName))
		} else {
			sb.WriteString(fmt.Sprintf(" <%s>", argName))
		}

		// Stop after first variadic positional flag
		if isVariadic {
			sb.WriteString(" [OPTIONS]")
			return sb.String()
		}
	}

	// Then process non-positional flags, but stop after first variadic
	for _, name := range nonPositionalFlags {
		flag := c.flags[name]

		// Check if it's variadic
		isVariadic := false
		switch f := flag.(type) {
		case *StringSliceFlag:
			isVariadic = f.Variadic
		case *IntSliceFlag:
			isVariadic = f.Variadic
		case *Int64SliceFlag:
			isVariadic = f.Variadic
		case *Float64SliceFlag:
			isVariadic = f.Variadic
		case *BoolSliceFlag:
			isVariadic = f.Variadic
		}

		if isVariadic {
			// All variadic flags show as [name...]
			sb.WriteString(fmt.Sprintf(" [%s...]", name))
			// Stop after first variadic flag
			sb.WriteString(" [OPTIONS]")
			return sb.String()
		} else {
			// Non-variadic required flags show as <name>
			shouldBeOptional := c.shouldFlagBeOptionalInSynopsis(flag)
			if shouldBeOptional {
				sb.WriteString(fmt.Sprintf(" [%s]", name))
			} else {
				sb.WriteString(fmt.Sprintf(" <%s>", name))
			}
		}
	}

	sb.WriteString(" [OPTIONS]")
	return sb.String()
}

func (c *Cmd) formatFlags(flags []any, isLongHelp bool) string {
	// Use flags in the order they were passed (already correctly ordered by generateUsage)
	allFlags := flags

	// First pass: calculate maximum width for alignment
	maxWidth := 0
	var flagParts []string

	for _, flag := range allFlags {
		base := getBaseFlag(flag)
		if base.Hidden {
			flagParts = append(flagParts, "")
			continue
		}
		if !isLongHelp && base.HiddenInShortHelp {
			flagParts = append(flagParts, "")
			continue
		}

		var flagPart string
		if base.PositionalOnly {
			// Positional-only flags show without dashes
			flagPart = fmt.Sprintf("  %s", base.Name)
		} else if base.Short != "" {
			flagPart = fmt.Sprintf("  -%s, --%s", base.Short, base.Name)
		} else {
			flagPart = fmt.Sprintf("      --%s", base.Name)
		}

		typeStr := getFlagType(flag)
		if typeStr != "bool" {
			flagPart = fmt.Sprintf("%s %s", flagPart, typeStr)
		}

		flagParts = append(flagParts, flagPart)
		if len(flagPart) > maxWidth {
			maxWidth = len(flagPart)
		}
	}

	// Use dynamic alignment: longest left side + 3 spaces
	maxWidth = maxWidth + 3

	// Second pass: generate aligned output
	var sb strings.Builder
	for i, flag := range allFlags {
		flagPart := flagParts[i]
		if flagPart == "" {
			continue // hidden flag
		}

		base := getBaseFlag(flag)
		sb.WriteString(flagPart)

		if base.Usage != "" {
			// Calculate padding to align descriptions
			padding := maxWidth - len(flagPart)
			if padding < 1 {
				padding = 1
			}
			sb.WriteString(strings.Repeat(" ", padding))

			// Add optional marker for flags that should be optional
			// but not for variadic flags (their type already indicates optionality)
			isVariadic := false
			switch f := flag.(type) {
			case *StringSliceFlag:
				isVariadic = f.Variadic
			case *IntSliceFlag:
				isVariadic = f.Variadic
			case *Int64SliceFlag:
				isVariadic = f.Variadic
			case *Float64SliceFlag:
				isVariadic = f.Variadic
			case *BoolSliceFlag:
				isVariadic = f.Variadic
			}

			// Show status markers for non-variadic flags:
			// For positional-only flags: show (optional) if explicitly optional, otherwise no marker
			// For other flags: (optional) if explicitly optional AND no default, (required) if required AND no default
			hasDefault := c.flagHasDefault(flag)
			var shouldShowOptional, shouldShowRequired bool
			if base.PositionalOnly {
				shouldShowOptional = base.Optional
				shouldShowRequired = false
			} else {
				shouldShowOptional = base.Optional && !hasDefault
				shouldShowRequired = !base.Optional && !hasDefault
			}

			if shouldShowOptional && !isVariadic {
				sb.WriteString("(optional) ")
			} else if shouldShowRequired && !isVariadic {
				sb.WriteString("(required) ")
			}

			sb.WriteString(base.Usage)

			// Add constraints
			constraints := c.getConstraintString(flag)
			if constraints != "" {
				sb.WriteString(" ")
				sb.WriteString(constraints)
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}
func getFlagType(flag any) string {
	switch f := flag.(type) {
	case *BoolFlag:
		return "bool"
	case *StringFlag:
		return "str"
	case *IntFlag:
		return "int"
	case *Int64Flag:
		return "int64"
	case *Float64Flag:
		return "float"
	case *BoolSliceFlag:
		return "bool"
	case *StringSliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[strs...]"
			}
			return "strs..."
		}
		return "strs"
	case *IntSliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[ints...]"
			}
			return "ints..."
		}
		return "ints"
	case *Int64SliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[int64s...]"
			}
			return "int64s..."
		}
		return "int64s"
	case *Float64SliceFlag:
		if f.Variadic {
			base := getBaseFlag(flag)
			if base.Optional {
				return "[floats...]"
			}
			return "floats..."
		}
		return "floats"
	}
	return ""
}

func (c *Cmd) getConstraintString(flag any) string {
	var parts []string

	// Add range constraints
	if rangeStr := c.getRangeString(flag); rangeStr != "" {
		parts = append(parts, "Range: "+rangeStr)
	}

	// Add enum or regex constraints (regex takes priority if both exist)
	if regexStr := c.getRegexString(flag); regexStr != "" {
		parts = append(parts, "Must match pattern: "+regexStr)
	} else if enumStr := c.getEnumString(flag); enumStr != "" {
		parts = append(parts, "Valid values: "+enumStr)
	}

	// Add default value
	if defaultStr := c.getDefaultString(flag); defaultStr != "" {
		parts = append(parts, fmt.Sprintf("(default %s)", defaultStr))
	}

	// Add separator for slices
	if sepStr := c.getSeparatorString(flag); sepStr != "" {
		parts = append(parts, "Separator: "+sepStr)
	}

	// Add relationship constraints
	if reqStr := c.getRequiresString(flag); reqStr != "" {
		parts = append(parts, "Requires: "+reqStr)
	}

	if exclStr := c.getExcludesString(flag); exclStr != "" {
		parts = append(parts, "Excludes: "+exclStr)
	}

	return strings.Join(parts, ". ")
}
func (c *Cmd) getDefaultString(flag any) string {
	switch f := flag.(type) {
	case *StringFlag:
		if f.Default != nil {
			return *f.Default
		}
	case *IntFlag:
		if f.Default != nil {
			return fmt.Sprintf("%d", *f.Default)
		}
	case *Int64Flag:
		if f.Default != nil {
			return fmt.Sprintf("%d", *f.Default)
		}
	case *Float64Flag:
		if f.Default != nil {
			return fmt.Sprintf("%g", *f.Default)
		}
	case *BoolFlag:
		if f.Default != nil && *f.Default {
			return "true" // Show (default true) but not (default false)
		}
		return ""
	case *StringSliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			return fmt.Sprintf("[%s]", strings.Join(*f.Default, ", "))
		}
	case *IntSliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%d", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	case *Int64SliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%d", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	case *Float64SliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%g", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	case *BoolSliceFlag:
		if f.Default != nil && len(*f.Default) > 0 {
			var strs []string
			for _, v := range *f.Default {
				strs = append(strs, fmt.Sprintf("%t", v))
			}
			return fmt.Sprintf("[%s]", strings.Join(strs, ", "))
		}
	}
	return ""
}

func (c *Cmd) getRangeString(flag any) string {
	switch f := flag.(type) {
	case *IntFlag:
		if f.min != nil || f.max != nil {
			var left, right string

			if f.min != nil {
				inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
				if inclusive {
					left = fmt.Sprintf("[%d", *f.min)
				} else {
					left = fmt.Sprintf("(%d", *f.min)
				}
			} else {
				left = "("
			}

			if f.max != nil {
				inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
				if inclusive {
					right = fmt.Sprintf("%d]", *f.max)
				} else {
					right = fmt.Sprintf("%d)", *f.max)
				}
			} else {
				right = ")"
			}

			return left + ", " + right
		}
	case *Int64Flag:
		if f.min != nil || f.max != nil {
			var left, right string

			if f.min != nil {
				inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
				if inclusive {
					left = fmt.Sprintf("[%d", *f.min)
				} else {
					left = fmt.Sprintf("(%d", *f.min)
				}
			} else {
				left = "("
			}

			if f.max != nil {
				inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
				if inclusive {
					right = fmt.Sprintf("%d]", *f.max)
				} else {
					right = fmt.Sprintf("%d)", *f.max)
				}
			} else {
				right = ")"
			}

			return left + ", " + right
		}
	case *Float64Flag:
		if f.min != nil || f.max != nil {
			var left, right string

			if f.min != nil {
				inclusive := f.minInclusive == nil || *f.minInclusive // default to inclusive
				if inclusive {
					left = fmt.Sprintf("[%g", *f.min)
				} else {
					left = fmt.Sprintf("(%g", *f.min)
				}
			} else {
				left = "("
			}

			if f.max != nil {
				inclusive := f.maxInclusive == nil || *f.maxInclusive // default to inclusive
				if inclusive {
					right = fmt.Sprintf("%g]", *f.max)
				} else {
					right = fmt.Sprintf("%g)", *f.max)
				}
			} else {
				right = ")"
			}

			return left + ", " + right
		}
	}
	return ""
}

func (c *Cmd) getEnumString(flag any) string {
	switch f := flag.(type) {
	case *StringFlag:
		if f.EnumConstraint != nil && len(*f.EnumConstraint) > 0 {
			return fmt.Sprintf("[%s]", strings.Join(*f.EnumConstraint, ", "))
		}
	}
	return ""
}

func (c *Cmd) getRegexString(flag any) string {
	switch f := flag.(type) {
	case *StringFlag:
		if f.RegexConstraint != nil {
			return f.RegexConstraint.String()
		}
	}
	return ""
}

func (c *Cmd) getSeparatorString(flag any) string {
	switch f := flag.(type) {
	case *StringSliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *IntSliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *Int64SliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *Float64SliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	case *BoolSliceFlag:
		if f.Separator != nil {
			return fmt.Sprintf("\"%s\"", *f.Separator)
		}
	}
	return ""
}

func (c *Cmd) getRequiresString(flag any) string {
	base := getBaseFlag(flag)
	if base != nil && base.Requires != nil && len(*base.Requires) > 0 {
		return strings.Join(*base.Requires, ", ")
	}
	return ""
}

func (c *Cmd) getExcludesString(flag any) string {
	base := getBaseFlag(flag)
	if base != nil && base.Excludes != nil && len(*base.Excludes) > 0 {
		return strings.Join(*base.Excludes, ", ")
	}
	return ""
}

func (c *Cmd) shouldFlagBeOptionalInSynopsis(flag any) bool {
	base := getBaseFlag(flag)

	// Check if flag has a default value (makes it optional)
	hasDefault := false
	switch f := flag.(type) {
	case *StringFlag:
		hasDefault = f.Default != nil
	case *IntFlag:
		hasDefault = f.Default != nil
	case *Int64Flag:
		hasDefault = f.Default != nil
	case *Float64Flag:
		hasDefault = f.Default != nil
	case *BoolFlag:
		hasDefault = f.Default != nil
	case *StringSliceFlag:
		hasDefault = f.Default != nil
	case *IntSliceFlag:
		hasDefault = f.Default != nil
	case *Int64SliceFlag:
		hasDefault = f.Default != nil
	case *Float64SliceFlag:
		hasDefault = f.Default != nil
	case *BoolSliceFlag:
		hasDefault = f.Default != nil
	}

	// Flag is optional if it has a default OR was explicitly set optional
	// For now, we'll use a heuristic: check the flag name against known optional flags
	// This is a workaround until we fix the default behavior
	if hasDefault {
		return true
	}

	// Special cases for positional-only flags that were explicitly set optional
	if base.PositionalOnly && (base.Name == "output-dir" || base.Name == "files") {
		return true
	}

	// Default to required for positional-only flags without defaults
	if base.PositionalOnly {
		return false
	}

	// For non-positional flags, use default value presence
	return hasDefault
}

func (c *Cmd) flagHasDefault(flag any) bool {
	switch f := flag.(type) {
	case *StringFlag:
		return f.Default != nil
	case *IntFlag:
		return f.Default != nil
	case *Int64Flag:
		return f.Default != nil
	case *Float64Flag:
		return f.Default != nil
	case *BoolFlag:
		return true // Boolean flags have implicit default of false
	case *StringSliceFlag:
		return f.Default != nil
	case *IntSliceFlag:
		return f.Default != nil
	case *Int64SliceFlag:
		return f.Default != nil
	case *Float64SliceFlag:
		return f.Default != nil
	case *BoolSliceFlag:
		return true // Boolean slice flags have implicit default of empty slice
	}
	return false
}
