package ra

type FlagSet struct {
	flags      map[string]*Flag[any]
	positional []string
}

func NewFlagSet() *FlagSet {
	return &FlagSet{
		flags:      make(map[string]*Flag[any]),
		positional: []string{},
	}
}

func (fs *FlagSet) Parse(args []string) *ParseError {
	return nil
}

// BASE TYPES

type BaseFlag struct {
	Name     string
	Short    string
	Usage    string
	Optional bool
	Hidden   bool
}

type Flag[T any] struct {
	BaseFlag
	Default *T
	Value   *T
}

// NON-SLICE FLAGS

type BoolFlag struct {
	Flag[bool]
}
type StringFlag struct {
	Flag[string]
}
type IntFlag struct {
	Flag[int]
	min *int
	max *int
}

// SLICE FLAGS

type SliceFlag[T any] struct {
	BaseFlag
	Separator *string
	Variadic  bool
	Default   *[]T
	Value     *[]T
}

type StringSliceFlag = SliceFlag[string]
type IntSliceFlag = SliceFlag[int]

// ADD METHODS

func (fs *FlagSet) AddBool(name string) *BoolFlag {
	def := false
	f := &BoolFlag{
		Flag: Flag[bool]{
			BaseFlag: BaseFlag{Name: name},
			Default:  &def,
		},
	}
	return f
}

func (fs *FlagSet) AddString(name string) *StringFlag {
	f := &StringFlag{
		Flag: Flag[string]{
			BaseFlag: BaseFlag{Name: name},
		},
	}
	return f
}

func (fs *FlagSet) AddInt(name string) *IntFlag {
	def := 0
	f := &IntFlag{
		Flag: Flag[int]{
			BaseFlag: BaseFlag{Name: name},
			Default:  &def,
		},
	}
	return f
}

func (fs *FlagSet) AddStringSlice(name string) *StringSliceFlag {
	f := &SliceFlag[string]{BaseFlag: BaseFlag{Name: name}}
	return f
}

func (fs *FlagSet) AddIntSlice(name string) *IntSliceFlag {
	f := &SliceFlag[int]{BaseFlag: BaseFlag{Name: name}}
	return f
}

// NON-SLICE SETTERS

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

func (f *IntFlag) SetMin(min int) *IntFlag {
	f.min = &min
	return f
}

func (f *IntFlag) SetMax(max int) *IntFlag {
	f.max = &max
	return f
}

// SLICE SETTERS

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

func (f *SliceFlag[T]) SetSeparator(sep string) *SliceFlag[T] {
	f.Separator = &sep
	return f
}

func (f *SliceFlag[T]) SetVariadic(b bool) *SliceFlag[T] {
	f.Variadic = b
	return f
}
