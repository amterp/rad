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

func (fs *FlagSet) Parse(args []string) error {
	return nil
}

// NON-SLICE FLAGS

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

type BoolFlag = Flag[bool]
type StringFlag = Flag[string]

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

// ADD

func (fs *FlagSet) AddBool(name string) *BoolFlag {
	def := false
	f := &BoolFlag{
		BaseFlag: BaseFlag{
			Name: name,
		},
		Default: &def,
	}
	return f
}

func (fs *FlagSet) AddString(name string) *StringFlag {
	f := &StringFlag{
		BaseFlag: BaseFlag{
			Name: name,
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

func (f *Flag[T]) SetShort(s string) *Flag[T] {
	f.Short = s
	return f
}

func (f *Flag[T]) SetUsage(u string) *Flag[T] {
	f.Usage = u
	return f
}

func (f *Flag[T]) SetDefault(v T) *Flag[T] {
	f.Default = &v
	return f
}

func (f *Flag[T]) SetOptional(b bool) *Flag[T] {
	f.Optional = b
	return f
}

func (f *Flag[T]) SetHidden(b bool) *Flag[T] {
	f.Hidden = b
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
