package ra

import "fmt"

type IntFlag struct {
	Flag[int]
	min          *int
	max          *int
	minInclusive *bool
	maxInclusive *bool
}

func NewInt(name string) *IntFlag {
	return &IntFlag{Flag: Flag[int]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func (f *IntFlag) SetShort(s string) *IntFlag {
	f.Short = s
	return f
}

func (f *IntFlag) SetUsage(u string) *IntFlag {
	f.Usage = u
	return f
}

func (f *IntFlag) SetDefault(v int) *IntFlag {
	f.Default = &v
	return f
}

func (f *IntFlag) SetOptional(b bool) *IntFlag {
	f.Optional = b
	return f
}

func (f *IntFlag) SetHidden(b bool) *IntFlag {
	f.Hidden = b
	return f
}

func (f *IntFlag) SetHiddenInShortHelp(b bool) *IntFlag {
	f.HiddenInShortHelp = b
	return f
}

func (f *IntFlag) SetPositionalOnly(b bool) *IntFlag {
	f.PositionalOnly = b
	return f
}

func (f *IntFlag) SetFlagOnly(b bool) *IntFlag {
	f.FlagOnly = b
	return f
}

func (f *IntFlag) SetExcludes(flags []string) *IntFlag {
	f.Excludes = &flags
	return f
}

func (f *IntFlag) SetRequires(flags []string) *IntFlag {
	f.Requires = &flags
	return f
}

func (f *IntFlag) SetMin(min int, inclusive bool) *IntFlag {
	f.min = &min
	f.minInclusive = &inclusive
	return f
}

func (f *IntFlag) SetMax(max int, inclusive bool) *IntFlag {
	f.max = &max
	f.maxInclusive = &inclusive
	return f
}

func (f *IntFlag) Register(cmd *Cmd, opts ...RegisterOption) (*int, error) {
	ptr := new(int)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *IntFlag) RegisterWithPtr(cmd *Cmd, ptr *int, opts ...RegisterOption) error {
	regConf := &registerConfig{}
	for _, opt := range opts {
		opt(regConf)
	}

	if _, exists := cmd.flags[f.Name]; exists {
		return fmt.Errorf("flag %q already defined", f.Name)
	}

	if regConf.global {
		cmd.globalFlags = append(cmd.globalFlags, f.Name)
	}

	// Create copy and set value pointer
	flag := *f
	flag.Value = ptr

	// Global flags should be flag-only (not positional)
	if regConf.global {
		flag.FlagOnly = true
	}

	// Add to short mapping
	if f.Short != "" {
		if _, exists := cmd.shortToName[f.Short]; exists {
			return fmt.Errorf("short flag %q already defined", f.Short)
		}
		cmd.shortToName[f.Short] = f.Name
	}

	cmd.flags[f.Name] = &flag
	if !flag.FlagOnly {
		// Check for positional-only after variadic error
		if flag.PositionalOnly {
			if err := cmd.validatePositionalOnlyAfterVariadic(f.Name); err != nil {
				return err
			}
		}
		cmd.positional = append(cmd.positional, f.Name)
	} else {
		cmd.nonPositional = append(cmd.nonPositional, f.Name)
	}

	return nil
}
