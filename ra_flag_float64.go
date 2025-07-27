package ra

import "fmt"

type Float64Flag struct {
	Flag[float64]
	min          *float64
	max          *float64
	minInclusive *bool
	maxInclusive *bool
}

func NewFloat64(name string) *Float64Flag {
	return &Float64Flag{Flag: Flag[float64]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}
func (f *Float64Flag) SetShort(s string) *Float64Flag {
	f.Short = s
	return f
}

func (f *Float64Flag) SetUsage(u string) *Float64Flag {
	f.Usage = u
	return f
}

func (f *Float64Flag) SetDefault(v float64) *Float64Flag {
	f.Default = &v
	return f
}

func (f *Float64Flag) SetOptional(b bool) *Float64Flag {
	f.Optional = b
	return f
}

func (f *Float64Flag) SetHidden(b bool) *Float64Flag {
	f.Hidden = b
	return f
}

func (f *Float64Flag) SetHiddenInLongHelp(b bool) *Float64Flag {
	f.HiddenInLongHelp = b
	return f
}

func (f *Float64Flag) SetPositionalOnly(b bool) *Float64Flag {
	f.PositionalOnly = b
	return f
}

func (f *Float64Flag) SetFlagOnly(b bool) *Float64Flag {
	f.FlagOnly = b
	return f
}

func (f *Float64Flag) SetExcludes(flags []string) *Float64Flag {
	f.Excludes = &flags
	return f
}

func (f *Float64Flag) SetRequires(flags []string) *Float64Flag {
	f.Requires = &flags
	return f
}

func (f *Float64Flag) SetMin(min float64, inclusive bool) *Float64Flag {
	f.min = &min
	f.minInclusive = &inclusive
	return f
}

func (f *Float64Flag) SetMax(max float64, inclusive bool) *Float64Flag {
	f.max = &max
	f.maxInclusive = &inclusive
	return f
}

func (f *Float64Flag) Register(cmd *Cmd, opts ...RegisterOption) (*float64, error) {
	ptr := new(float64)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *Float64Flag) RegisterWithPtr(cmd *Cmd, ptr *float64, opts ...RegisterOption) error {
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
