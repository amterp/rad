package core

import (
	"fmt"
	"github.com/spf13/pflag"
	"strconv"
	"strings"
)

type RslArg interface {
	GetName() string
	GetIdentifier() string
	GetShort() string
	GetArgUsage() string
	GetDescription() string
	DefaultAsString() string
	HasNonZeroDefault() bool
	isRegistered() bool
	Register()
	Configured() bool
	Lookup() *pflag.Flag
	SetValue(value string)
	IsOptional() bool
	GetToken() Token // nil if not a script arg
	Hidden(bool)
	IsHidden() bool
}

type BaseRslArg struct {
	Name              string // User-facing arg they'll set in CLI
	Identifier        string // Identifier in script. If non-script arg, then equal to Name
	Short             string
	ArgUsage          string
	Description       string
	defaultAsString   string
	hasNonZeroDefault bool
	registered        bool
	scriptArg         *ScriptArg
	hidden            bool
}

func (f *BaseRslArg) GetName() string {
	return f.Name
}

func (f *BaseRslArg) GetIdentifier() string {
	return f.Identifier
}

func (f *BaseRslArg) GetShort() string {
	return f.Short
}

func (f *BaseRslArg) GetArgUsage() string {
	return f.ArgUsage
}

func (f *BaseRslArg) GetDescription() string {
	return f.Description
}

func (f *BaseRslArg) DefaultAsString() string {
	return f.defaultAsString
}

func (f *BaseRslArg) HasNonZeroDefault() bool {
	return f.hasNonZeroDefault
}

func (f *BaseRslArg) isRegistered() bool {
	return f.registered
}

func (f *BaseRslArg) Configured() bool {
	return pflag.Lookup(f.Name).Changed
}

func (f *BaseRslArg) Lookup() *pflag.Flag {
	return pflag.Lookup(f.Name)
}

func (f *BaseRslArg) IsOptional() bool {
	if f.scriptArg == nil {
		return true
	}

	return f.scriptArg.IsOptional
}

func (f *BaseRslArg) GetToken() Token {
	if f.scriptArg == nil {
		return nil
	}

	return f.scriptArg.DeclarationToken
}

func (f *BaseRslArg) Hidden(hide bool) {
	f.hidden = hide
}

func (f *BaseRslArg) IsHidden() bool {
	return f.hidden
}

// --- bool

type BoolRslFlag struct {
	BaseRslArg
	Value   bool
	Default bool
}

func NewBoolRadFlag(name, short, description string, defaultValue bool) BoolRslFlag {
	return BoolRslFlag{
		BaseRslArg: BaseRslArg{
			Name:              name,
			Identifier:        name,
			Short:             short,
			ArgUsage:          "",
			Description:       description,
			defaultAsString:   fmt.Sprint(defaultValue),
			hasNonZeroDefault: defaultValue != false,
		},
		Default: defaultValue,
	}
}

func (f *BoolRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.BoolVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *BoolRslFlag) SetValue(arg string) {
	arg = strings.ToLower(arg)
	if arg == "true" || arg == "1" {
		f.Value = true
	} else if arg == "false" || arg == "0" {
		f.Value = false
	} else {
		RP.TokenErrorExit(f.scriptArg.DeclarationToken, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg))
	}
}

// --- bool array

type BoolArrRslFlag struct {
	BaseRslArg
	Value   []bool
	Default []bool
}

func NewBoolArrRadFlag(name, short, argUsage, description string, defaultValue []bool) BoolArrRslFlag {
	return BoolArrRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *BoolArrRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.BoolSliceVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *BoolArrRslFlag) SetValue(arg string) {
	// split on arg commas
	split := strings.Split(arg, ",")
	bools := make([]bool, len(split))
	for i, v := range split {
		v = strings.ToLower(v)
		if v == "true" || v == "1" {
			bools[i] = true
		} else if v == "false" || v == "0" {
			bools[i] = false
		} else {
			RP.TokenErrorExit(f.scriptArg.DeclarationToken, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg))
		}
	}
	f.Value = bools
}

// --- string

type StringRslFlag struct {
	BaseRslArg
	Value   string
	Default string
}

func NewStringRadFlag(name, short, argUsage, description, defaultValue string) StringRslFlag {
	return StringRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   defaultValue,
			hasNonZeroDefault: defaultValue != "",
		},
		Default: defaultValue,
	}
}

func (f *StringRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.StringVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *StringRslFlag) SetValue(arg string) {
	f.Value = arg
}

// --- string array

type StringArrRslFlag struct {
	BaseRslArg
	Value   []string
	Default []string
}

func NewStringArrRadFlag(name, short, argUsage, description string, defaultValue []string) StringArrRslFlag {
	return StringArrRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *StringArrRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.StringArrayVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *StringArrRslFlag) SetValue(arg string) {
	// split on arg commas
	split := strings.Split(arg, ",")
	vals := make([]string, len(split))
	for i, v := range split {
		vals[i] = v
	}
	f.Value = vals
}

// --- int

type IntRslFlag struct {
	BaseRslArg
	Value   int64
	Default int64
}

func NewIntRadFlag(name, short, argUsage, description string, defaultValue int64) IntRslFlag {
	return IntRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: defaultValue != 0,
		},
		Default: defaultValue,
	}
}

func (f *IntRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.Int64VarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *IntRslFlag) SetValue(arg string) {
	parsed, err := strconv.Atoi(arg)
	if err != nil {
		RP.TokenErrorExit(f.scriptArg.DeclarationToken, fmt.Sprintf("Expected int, but could not parse: %v\n", arg))
	}
	val := int64(parsed)
	f.Value = val
}

// --- int array

type IntArrRslFlag struct {
	BaseRslArg
	Value   []int64
	Default []int64
}

func NewIntArrRadFlag(name, short, argUsage, description string, defaultValue []int64) IntArrRslFlag {
	return IntArrRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *IntArrRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.Int64SliceVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *IntArrRslFlag) SetValue(arg string) {
	// split on arg commas
	split := strings.Split(arg, ",")
	ints := make([]int64, len(split))
	for i, v := range split {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			RP.TokenErrorExit(f.scriptArg.DeclarationToken, fmt.Sprintf("Expected int, but could not parse: %v\n", arg))
		}
		ints[i] = parsed
	}
	f.Value = ints
}

// --- float

type FloatRslFlag struct {
	BaseRslArg
	Value   float64
	Default float64
}

func NewFloatRadFlag(name, short, argUsage, description string, defaultValue float64) FloatRslFlag {
	return FloatRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: defaultValue != 0,
		},
		Default: defaultValue,
	}
}

func (f *FloatRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.Float64VarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *FloatRslFlag) SetValue(arg string) {
	parsed, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		RP.TokenErrorExit(f.scriptArg.DeclarationToken, fmt.Sprintf("Expected float, but could not parse: %v\n", arg))
	}
	f.Value = parsed
}

// --- float array

type FloatArrRslFlag struct {
	BaseRslArg
	Value   []float64
	Default []float64
}

func NewFloatArrRadFlag(name, short, argUsage, description string, defaultValue []float64) FloatArrRslFlag {
	return FloatArrRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(defaultValue),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *FloatArrRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.Float64SliceVarP(&f.Value, f.Name, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *FloatArrRslFlag) SetValue(arg string) {
	// split on arg commas
	split := strings.Split(arg, ",")
	floats := make([]float64, len(split))
	for i, v := range split {
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			RP.TokenErrorExit(f.scriptArg.DeclarationToken, fmt.Sprintf("Expected float, but could not parse: %v\n", arg))
		}
		floats[i] = parsed
	}
	f.Value = floats
}

// --- MockResponse

type MockResponseRslFlag struct {
	BaseRslArg
	Value MockResponseSlice
}

func NewMockResponseRadFlag(name, short, usage string) MockResponseRslFlag {
	return MockResponseRslFlag{
		BaseRslArg: BaseRslArg{
			Name: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          "string",
			Description:       usage,
			defaultAsString:   "",
			hasNonZeroDefault: false,
		},
	}
}

func (f *MockResponseRslFlag) Register() {
	if f.registered {
		return
	}

	pflag.VarP(&f.Value, f.Name, f.Short, f.Description)

	f.registered = true
}

func (f *MockResponseRslFlag) SetValue(arg string) {
	RP.RadErrorExit(fmt.Sprintf("This function is expected to only be called for script args."+
		" MockResponse cannot be a script arg: %v\n", arg))
}
