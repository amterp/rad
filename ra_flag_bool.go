package ra

import "fmt"

type BoolFlag struct {
	Flag[bool]
}

func NewBool(name string) *BoolFlag {
	return &BoolFlag{Flag: Flag[bool]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}

func (f *BoolFlag) SetShort(s string) *BoolFlag {
	f.Short = s
	return f
}

func (f *BoolFlag) SetUsage(u string) *BoolFlag {
	f.Usage = u
	return f
}

func (f *BoolFlag) SetDefault(v bool) *BoolFlag {
	f.Default = &v
	return f
}

func (f *BoolFlag) SetOptional(b bool) *BoolFlag {
	f.Optional = b
	return f
}

func (f *BoolFlag) SetHidden(b bool) *BoolFlag {
	f.Hidden = b
	return f
}

func (f *BoolFlag) SetHiddenInLongHelp(b bool) *BoolFlag {
	f.HiddenInLongHelp = b
	return f
}

func (f *BoolFlag) SetPositionalOnly(b bool) *BoolFlag {
	f.PositionalOnly = b
	return f
}

func (f *BoolFlag) SetFlagOnly(b bool) *BoolFlag {
	f.FlagOnly = b
	return f
}

func (f *BoolFlag) SetExcludes(flags []string) *BoolFlag {
	f.Excludes = &flags
	return f
}

func (f *BoolFlag) SetRequires(flags []string) *BoolFlag {
	f.Requires = &flags
	return f
}

func (f *BoolFlag) Register(cmd *Cmd, opts ...RegisterOption) (*bool, error) {
	ptr := new(bool)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *BoolFlag) RegisterWithPtr(cmd *Cmd, ptr *bool, opts ...RegisterOption) error {
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
	// Bool flags are always flag-only (non-positional)
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

	// Bool flags are always non-positional (flag-only)
	cmd.nonPositional = append(cmd.nonPositional, f.Name)

	return nil
}
