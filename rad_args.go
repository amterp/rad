package ra

type FlagSet struct{}

func NewFlagSet() *FlagSet {
	return &FlagSet{}
}

//

type BoolFlag = Flag[bool]
type StringFlag = Flag[string]

//

func (fs *FlagSet) Parse(args []string) error {
	return nil
}

//

func (fs *FlagSet) AddBool(name string) *BoolFlag {
	return &Flag[bool]{
		Name:    name,
		Default: false,
	}
}

func (fs *FlagSet) AddString(name string) *StringFlag {
	return &Flag[string]{Name: name}
}

//

type Flag[T any] struct {
	Name     string
	Short    string
	Usage    string
	Optional bool
	Default  T
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
	f.Default = v
	f.Value = &v
	return f
}

func (f *Flag[T]) SetOptional(b bool) *Flag[T] {
	f.Optional = b
	return f
}
