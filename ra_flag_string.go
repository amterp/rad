package ra

import (
	"fmt"
	"regexp"
)

type StringFlag struct {
	Flag[string]
	EnumConstraint  *[]string      // if set, the value must be one of these
	RegexConstraint *regexp.Regexp // if set, the value must match this regex
}

func NewString(name string) *StringFlag {
	return &StringFlag{Flag: Flag[string]{BaseFlag: BaseFlag{Name: name, Optional: false}}}
}
func (f *StringFlag) SetShort(s string) *StringFlag {
	f.Short = s
	return f
}

func (f *StringFlag) SetUsage(u string) *StringFlag {
	f.Usage = u
	return f
}

func (f *StringFlag) SetDefault(v string) *StringFlag {
	f.Default = &v
	return f
}

func (f *StringFlag) SetOptional(b bool) *StringFlag {
	f.Optional = b
	return f
}

func (f *StringFlag) SetHidden(b bool) *StringFlag {
	f.Hidden = b
	return f
}

func (f *StringFlag) SetHiddenInShortHelp(b bool) *StringFlag {
	f.HiddenInShortHelp = b
	return f
}

func (f *StringFlag) SetPositionalOnly(b bool) *StringFlag {
	f.PositionalOnly = b
	return f
}

func (f *StringFlag) SetFlagOnly(b bool) *StringFlag {
	f.FlagOnly = b
	return f
}

func (f *StringFlag) SetExcludes(flags []string) *StringFlag {
	f.Excludes = &flags
	return f
}

func (f *StringFlag) SetRequires(flags []string) *StringFlag {
	f.Requires = &flags
	return f
}

func (f *StringFlag) SetEnumConstraint(values []string) *StringFlag {
	if len(values) == 0 {
		f.EnumConstraint = nil
	} else {
		f.EnumConstraint = &values
	}
	return f
}

func (f *StringFlag) SetRegexConstraint(regex *regexp.Regexp) *StringFlag {
	f.RegexConstraint = regex
	return f
}

func (f *StringFlag) Register(cmd *Cmd, opts ...RegisterOption) (*string, error) {
	ptr := new(string)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *StringFlag) RegisterWithPtr(cmd *Cmd, ptr *string, opts ...RegisterOption) error {
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
		flag.Optional = true // TODO test, make clearer? Might be surprising/undesirable?
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
