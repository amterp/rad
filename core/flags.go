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
	IsSet() bool
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

func (f *BaseRadFlag) IsSet() bool {
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

func NewBoolRadFlag(name, short, usage string, defaultValue bool) BoolRadFlag {
	return BoolRadFlag{
		BaseRadFlag: BaseRadFlag{
			Name:              name,
			Short:             short,
			ArgUsage:          "",
			Description:       usage,
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

	pflag.BoolVar(&f.Value, f.Name, f.Default, f.Description)

	if f.Short != "" {
		pflag.Lookup(f.Name).Shorthand = f.Short
	}

	f.registered = true
}

// --- string

type StringRadFlag struct {
	BaseRadFlag
	Value   string
	Default string
}

func NewStringRadFlag(name, short, argUsage, description string, defaultValue string) StringRadFlag {
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

	pflag.StringVar(&f.Value, f.Name, f.Default, f.Description)

	if f.Short != "" {
		pflag.Lookup(f.Name).Shorthand = f.Short
	}

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

	pflag.Var(&f.Value, f.Name, f.Description)

	if f.Short != "" {
		pflag.Lookup(f.Name).Shorthand = f.Short
	}

	f.registered = true
}
