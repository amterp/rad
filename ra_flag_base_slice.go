package ra

import "fmt"

type SliceFlag[T any] struct {
	BaseFlag
	Separator *string
	Variadic  bool
	Default   *[]T
	Value     *[]T
}

type StringSliceFlag = SliceFlag[string]
type IntSliceFlag = SliceFlag[int]
type Int64SliceFlag = SliceFlag[int64]
type Float64SliceFlag = SliceFlag[float64]
type BoolSliceFlag = SliceFlag[bool]

func NewStringSlice(name string) *StringSliceFlag {
	return &SliceFlag[string]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewIntSlice(name string) *IntSliceFlag {
	return &SliceFlag[int]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewInt64Slice(name string) *Int64SliceFlag {
	return &SliceFlag[int64]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewFloat64Slice(name string) *Float64SliceFlag {
	return &SliceFlag[float64]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}

func NewBoolSlice(name string) *BoolSliceFlag {
	return &SliceFlag[bool]{BaseFlag: BaseFlag{Name: name, Optional: false}}
}
func (f *SliceFlag[T]) SetShort(s string) *SliceFlag[T] {
	f.Short = s
	return f
}

func (f *SliceFlag[T]) SetUsage(u string) *SliceFlag[T] {
	f.Usage = u
	return f
}

func (f *SliceFlag[T]) SetDefault(v []T) *SliceFlag[T] {
	f.Default = &v
	return f
}

func (f *SliceFlag[T]) SetOptional(b bool) *SliceFlag[T] {
	f.Optional = b
	return f
}

func (f *SliceFlag[T]) SetHidden(b bool) *SliceFlag[T] {
	f.Hidden = b
	return f
}

func (f *SliceFlag[T]) SetHiddenInLongHelp(b bool) *SliceFlag[T] {
	f.HiddenInLongHelp = b
	return f
}

func (f *SliceFlag[T]) SetPositionalOnly(b bool) *SliceFlag[T] {
	f.PositionalOnly = b
	return f
}

func (f *SliceFlag[T]) SetFlagOnly(b bool) *SliceFlag[T] {
	f.FlagOnly = b
	return f
}

func (f *SliceFlag[T]) SetExcludes(flags []string) *SliceFlag[T] {
	f.Excludes = &flags
	return f
}

func (f *SliceFlag[T]) SetRequires(flags []string) *SliceFlag[T] {
	f.Requires = &flags
	return f
}

func (f *SliceFlag[T]) SetSeparator(sep string) *SliceFlag[T] {
	f.Separator = &sep
	return f
}

func (f *SliceFlag[T]) SetVariadic(b bool) *SliceFlag[T] {
	f.Variadic = b
	return f
}

func (f *SliceFlag[T]) Register(cmd *Cmd, opts ...RegisterOption) (*[]T, error) {
	ptr := new([]T)
	return ptr, f.RegisterWithPtr(cmd, ptr, opts...)
}

func (f *SliceFlag[T]) RegisterWithPtr(cmd *Cmd, ptr *[]T, opts ...RegisterOption) error {
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
