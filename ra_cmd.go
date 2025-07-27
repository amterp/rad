package ra

import (
	"fmt"
)

type Cmd struct {
	name          string
	description   string
	flags         map[string]any // flag name -> flag itself (either a Flag[T] or SliceFlag[T])
	positional    []string       // positional flags, i.e. flags that are positional args
	nonPositional []string       // non-positional flags, i.e. flags that are only named
	globalFlags   []string       // flags that will be applied to all subcommands
	subCmds       map[string]*Cmd
	shortToName   map[string]string // short flag -> full name mapping

	// options
	customUsage          func(bool) // if set, this function will be called to print usage instead of the default
	helpEnabled          bool       // default true automatically adds a help flag
	excludeNameFromUsage bool       // if true, this command will not be included in usage output

	// state post-parse
	used             *bool           // after parsing, whether this command was invoked
	configured       map[string]bool // specified flags from flags.
	unknownArgs      []string        // unknown args when ignoreUnknown is true
	lastVariadicFlag string          // last variadic flag that was used
	sawFlag          bool            // true if we've seen a flag since the last variadic
}

func NewCmd(name string) *Cmd {
	c := &Cmd{
		name:        name,
		flags:       make(map[string]any),
		positional:  []string{},
		subCmds:     make(map[string]*Cmd),
		configured:  make(map[string]bool),
		helpEnabled: true,
		shortToName: make(map[string]string),
	}

	return c
}

func (c *Cmd) SetDescription(desc string) *Cmd {
	c.description = desc
	return c
}

func (c *Cmd) SetCustomUsage(fn func(isLongHelp bool)) *Cmd {
	c.customUsage = fn
	return c
}

func (c *Cmd) SetHelpEnabled(enable bool) *Cmd {
	c.helpEnabled = enable
	return c
}

func (c *Cmd) SetExcludeNameFromUsage(exclude bool) *Cmd {
	c.excludeNameFromUsage = exclude
	return c
}

func (c *Cmd) applyGlobalFlags(subCmd *Cmd) error {
	for _, globalFlagName := range c.globalFlags {
		if flag, exists := c.flags[globalFlagName]; exists {
			// Only add flag if it doesn't already exist in subcommand
			if _, exists := subCmd.flags[globalFlagName]; !exists {
				subCmd.flags[globalFlagName] = flag
				if base := getBaseFlag(flag); base != nil && base.Short != "" {
					subCmd.shortToName[base.Short] = base.Name
				}
				// Also add to subcommand's global flags list and non-positional list
				subCmd.globalFlags = append(subCmd.globalFlags, globalFlagName)
				subCmd.nonPositional = append(subCmd.nonPositional, globalFlagName)
			}
		}
	}
	return nil
}

// Whether a flag was explicitly configured by the user.
func (c *Cmd) Configured(name string) bool {
	// Check if flag is configured in this command
	if configured, exists := c.configured[name]; exists && configured {
		return true
	}

	// Check all invoked subcommands recursively
	for _, subCmd := range c.subCmds {
		if subCmd.used != nil && *subCmd.used {
			if subCmd.Configured(name) {
				return true
			}
		}
	}

	return false
}

func (c *Cmd) GetUnknownArgs() []string {
	return c.unknownArgs
}
func (c *Cmd) RegisterCmd(subCmd *Cmd) (*bool, error) {
	if _, exists := c.subCmds[subCmd.name]; exists {
		return nil, fmt.Errorf("command %q already defined", subCmd.name)
	}

	c.subCmds[subCmd.name] = subCmd
	subCmd.used = new(bool)

	// Apply global flags to subcommand for usage generation
	if err := c.applyGlobalFlags(subCmd); err != nil {
		return nil, err
	}

	return subCmd.used, nil
}

func (c *Cmd) validatePositionalOnlyAfterVariadic(flagName string) error {
	// Check if there's already a variadic positional flag
	for _, existingName := range c.positional {
		existingFlag := c.flags[existingName]

		// Check if this existing flag is variadic
		switch f := existingFlag.(type) {
		case *StringSliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *IntSliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *Int64SliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *Float64SliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		case *BoolSliceFlag:
			if f.Variadic {
				return fmt.Errorf("cannot register positional-only flag %q after variadic positional flag %q (positional-only flags cannot be set after variadic flags)", flagName, existingName)
			}
		}
	}
	return nil
}
