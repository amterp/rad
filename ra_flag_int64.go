package ra

import "fmt"

type Int64Flag struct {
	Flag[int64]
	min          *int64
	max          *int64
	minInclusive *bool
	maxInclusive *bool
}

func NewInt64(name string) *Int64Flag {
	return &Int64Flag{Flag: Flag[int64]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func (f *Int64Flag) SetShort(s string) *Int64Flag {
	f.Short = s
	return f
}

func (f *Int64Flag) SetUsage(u string) *Int64Flag {
	f.Usage = u
	return f
}

func (f *Int64Flag) SetDefault(v int64) *Int64Flag {
	f.Default = &v
	return f
}

func (f *Int64Flag) SetOptional(b bool) *Int64Flag {
	f.Optional = b
	return f
}

func (f *Int64Flag) SetHidden(b bool) *Int64Flag {
	f.Hidden = b
	return f
}

func (f *Int64Flag) SetHiddenInShortHelp(b bool) *Int64Flag {
	f.HiddenInShortHelp = b
	return f
}

func (f *Int64Flag) SetPositionalOnly(b bool) *Int64Flag {
	f.PositionalOnly = b
	return f
}

func (f *Int64Flag) SetFlagOnly(b bool) *Int64Flag {
	f.FlagOnly = b
	return f
}

func (f *Int64Flag) SetExcludes(flags []string) *Int64Flag {
	f.Excludes = &flags
	return f
}

func (f *Int64Flag) SetRequires(flags []string) *Int64Flag {
	f.Requires = &flags
	return f
}

func (f *Int64Flag) SetMin(min int64, inclusive bool) *Int64Flag {
	f.min = &min
	f.minInclusive = &inclusive
	return f
}

func (f *Int64Flag) SetMax(max int64, inclusive bool) *Int64Flag {
	f.max = &max
	f.maxInclusive = &inclusive
	return f
}

func (f *Int64Flag) Register(cmd *Cmd, opts ...RegisterOption) (*int64, error) {
	ptr := new(int64)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *Int64Flag) RegisterWithPtr(cmd *Cmd, ptr *int64, opts ...RegisterOption) error {
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
