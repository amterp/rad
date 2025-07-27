package ra

type BaseFlag struct {
	Name             string
	Short            string
	Usage            string
	Optional         bool
	Hidden           bool
	HiddenInLongHelp bool
	PositionalOnly   bool
	FlagOnly         bool
	Excludes         *[]string
	Requires         *[]string
}
type Flag[T any] struct {
	BaseFlag
	Default *T
	Value   *T
}
