package core

import (
	"fmt"
	"github.com/spf13/pflag"
)

type RadFlag interface {
	GetName() string
	GetShort() string
	GetArgUsage() string
	GetDescription() string
	DefaultAsString() string
	HasNonZeroDefault() bool
	isRegistered() bool
	Register()
	Configured() bool
	Lookup() *pflag.Flag
}

type BaseRadFlag struct {
	Name              string
	Short             string
	ArgUsage          string
	Description       string
	defaultAsString   string
	hasNonZeroDefault bool
	registered        bool
}

func (f *BaseRadFlag) GetName() string {
	return f.Name
}

func (f *BaseRadFlag) GetShort() string {
	return f.Short
}

func (f *BaseRadFlag) GetArgUsage() string {
	return f.ArgUsage
}

func (f *BaseRadFlag) GetDescription() string {
	return f.Description
}

func (f *BaseRadFlag) DefaultAsString() string {
	return f.defaultAsString
}

func (f *BaseRadFlag) HasNonZeroDefault() bool {
	return f.hasNonZeroDefault
}

func (f *BaseRadFlag) isRegistered() bool {
	return f.registered
}

func (f *BaseRadFlag) Configured() bool {
	return pflag.Lookup(f.Name).Changed
}

func (f *BaseRadFlag) Lookup() *pflag.Flag {
	return pflag.Lookup(f.Name)
}

// --- bool

type BoolRadFlag struct {
	BaseRadFlag
	Value   bool
	Default bool
}

func NewBoolRadFlag(name, short, description string, defaultValue bool) BoolRadFlag {
	return BoolRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          "",
			Description:       description,
			defaultAsString:   fmt.Sprint(defaultValue),
			hasNonZeroDefault: defaultValue != false,
		},
		Default: defaultValue,
	}
}

func (f *BoolRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.BoolVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- bool array

type BoolArrRadFlag struct {
	BaseRadFlag
	Value   []bool
	Default []bool
}

func NewBoolArrRadFlag(name, short, argUsage, description string, defaultValue []bool) BoolArrRadFlag {
	return BoolArrRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *BoolArrRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.BoolSliceVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- string

type StringRadFlag struct {
	BaseRadFlag
	Value   string
	Default string
}

func NewStringRadFlag(name, short, argUsage, description, defaultValue string) StringRadFlag {
	return StringRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   defaultValue,
			hasNonZeroDefault: defaultValue != "",
		},
		Default: defaultValue,
	}
}

func (f *StringRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.StringVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- string array

type StringArrRadFlag struct {
	BaseRadFlag
	Value   []string
	Default []string
}

func NewStringArrRadFlag(name, short, argUsage, description string, defaultValue []string) StringArrRadFlag {
	return StringArrRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *StringArrRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.StringArrayVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- int

type IntRadFlag struct {
	BaseRadFlag
	Value   int64
	Default int64
}

func NewIntRadFlag(name, short, argUsage, description string, defaultValue int64) IntRadFlag {
	return IntRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: defaultValue != 0,
		},
		Default: defaultValue,
	}
}

func (f *IntRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.Int64VarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- int array

type IntArrRadFlag struct {
	BaseRadFlag
	Value   []int64
	Default []int64
}

func NewIntArrRadFlag(name, short, argUsage, description string, defaultValue []int64) IntArrRadFlag {
	return IntArrRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *IntArrRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.Int64SliceVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- float

type FloatRadFlag struct {
	BaseRadFlag
	Value   float64
	Default float64
}

func NewFloatRadFlag(name, short, argUsage, description string, defaultValue float64) FloatRadFlag {
	return FloatRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: defaultValue != 0,
		},
		Default: defaultValue,
	}
}

func (f *FloatRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.Float64VarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- float array

type FloatArrRadFlag struct {
	BaseRadFlag
	Value   []float64
	Default []float64
}

func NewFloatArrRadFlag(name, short, argUsage, description string, defaultValue []float64) FloatArrRadFlag {
	return FloatArrRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *FloatArrRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.Float64SliceVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

// --- MockResponse

type MockResponseRadFlag struct {
	BaseRadFlag
	Value MockResponseSlice
}

func NewMockResponseRadFlag(name, short, usage string) MockResponseRadFlag {
	return MockResponseRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          "string",
			Description:       usage,
			defaultAsString:   "",
			hasNonZeroDefault: false,
		},
	}
}

func (f *MockResponseRadFlag) Register() {
	if f.registered {
		return
	}

	pflag.VarP(&f.Value, f.Name, f.Short, f.Description)

	f.registered = true
}
