package ra

type FlagSet struct{}

func NewFlagSet() *FlagSet {
	return &FlagSet{}
}

//

type BoolFlag = Flag[bool]
type StringFlag = Flag[string]
type StringSliceFlag = SliceFlag[string]

//

func (fs *FlagSet) Parse(args []string) error {
	return nil
}

//

func (fs *FlagSet) AddBool(name string) *BoolFlag {
	def := false
	return &BoolFlag{
		Name:    name,
		Default: &def,
	}
}

func (fs *FlagSet) AddString(name string) *StringFlag {
	return &StringFlag{Name: name}
}

func (fs *FlagSet) AddStringSlice(name string) *StringSliceFlag {
	return &StringSliceFlag{
		Name: name,
	}
}

//

type Flag[T any] struct {
	Name     string
	Short    string
	Usage    string
	Optional bool
	Hidden   bool
	Default  *T
	Value    *T
}

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

type SliceFlag[T any] struct {
	Name      string
	Short     string
	Usage     string
	Optional  bool
	Hidden    bool
	Separator *string
	Variadic  bool
	Default   *[]T
	Value     *[]T
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

func (f *SliceFlag[T]) SetSeparator(sep string) *SliceFlag[T] {
	f.Separator = &sep
	return f
}

func (f *SliceFlag[T]) SetVariadic(b bool) *SliceFlag[T] {
	f.Variadic = b
	return f
}
