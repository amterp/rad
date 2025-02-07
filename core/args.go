package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/samber/lo"
	"github.com/spf13/pflag"
)

type RslArg interface {
	GetExternalName() string
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
	GetNode() *ts.Node // nil if not a script arg
	Hidden(bool)
	IsHidden() bool
	ValidateConstraints() error
}

type BaseRslArg struct {
	ExternalName      string // User-facing arg they'll set in CLI
	Identifier        string // Identifier in script. If non-script arg, then equal to ExternalName
	Short             string
	ArgUsage          string
	Description       string
	defaultAsString   string
	hasNonZeroDefault bool
	registered        bool
	scriptArg         *ScriptArg
	hidden            bool
}

func (f *BaseRslArg) GetExternalName() string {
	return f.ExternalName
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
	return RFlagSet.Lookup(f.ExternalName).Changed
}

func (f *BaseRslArg) Lookup() *pflag.Flag {
	return RFlagSet.Lookup(f.ExternalName)
}

func (f *BaseRslArg) IsOptional() bool {
	if f.scriptArg == nil {
		return true
	}

	return f.scriptArg.IsOptional
}

func (f *BaseRslArg) GetNode() *ts.Node {
	if f.scriptArg == nil {
		return nil
	}

	return f.scriptArg.Decl.Node()
}

func (f *BaseRslArg) Hidden(hide bool) {
	f.hidden = hide
}

func (f *BaseRslArg) IsHidden() bool {
	return f.hidden
}

func (f *BaseRslArg) ValidateConstraints() error {
	// Base impl does nothing -- each arg type will implement its own constraints
	return nil
}

// --- bool

type BoolRslArg struct {
	BaseRslArg
	Value   bool
	Default bool
}

func NewBoolRadArg(name, short, description string, defaultValue bool) BoolRslArg {
	return BoolRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName:      name,
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

func (f *BoolRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.BoolVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *BoolRslArg) SetValue(arg string) {
	arg = strings.ToLower(arg)
	if arg == "true" || arg == "1" {
		f.Value = true
	} else if arg == "false" || arg == "0" {
		f.Value = false
	} else {
		RP.CtxErrorExit(NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg)))
	}
}

// --- bool array

type BoolArrRslArg struct {
	BaseRslArg
	Value   []bool
	Default []bool
}

func NewBoolArrRadArg(name, short, argUsage, description string, defaultValue []bool) BoolArrRslArg {
	return BoolArrRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName:      name,
			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *BoolArrRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.BoolSliceVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *BoolArrRslArg) SetValue(arg string) {
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
			RP.CtxErrorExit(NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg)))
		}
	}
	f.Value = bools
}

// --- string

type StringRslArg struct {
	BaseRslArg
	Value           string
	Default         string
	EnumConstraint  *[]string
	RegexConstraint *regexp.Regexp
}

func NewStringRadArg(name, short, argUsage, description, defaultValue string, enum *[]string, regex *regexp.Regexp) StringRslArg {
	return StringRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName:      name,
			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   defaultValue,
			hasNonZeroDefault: defaultValue != "",
		},
		Default:         defaultValue,
		EnumConstraint:  enum,
		RegexConstraint: regex,
	}
}

func (f *StringRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.StringVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *StringRslArg) SetValue(arg string) {
	f.Value = arg
}

func (f *StringRslArg) GetDescription() string {
	var builder strings.Builder

	builder.WriteString(f.Description)

	if f.EnumConstraint != nil {
		builder.WriteString(" Valid values: [")
		builder.WriteString(strings.Join(*f.EnumConstraint, ", "))
		builder.WriteString("].")
	}

	if f.RegexConstraint != nil {
		builder.WriteString(" Regex: ")
		builder.WriteString((*f.RegexConstraint).String())
	}

	return builder.String()
}

//goland:noinspection GoErrorStringFormat
func (f *StringRslArg) ValidateConstraints() error {
	if f.EnumConstraint != nil {
		if !lo.Contains(*f.EnumConstraint, f.Value) {
			return fmt.Errorf("Invalid '%s' value: %v (valid values: %s)", f.ExternalName, f.Value, strings.Join(*f.EnumConstraint, ", "))
		}
	}

	constraint := f.RegexConstraint
	if constraint != nil {
		if !constraint.MatchString(f.Value) {
			return fmt.Errorf("Invalid '%s' value: %v (must match regex: %s)", f.ExternalName, f.Value, constraint.String())
		}
	}

	return nil
}

// --- string array

type StringArrRslArg struct {
	BaseRslArg
	Value   []string
	Default []string
}

func NewStringArrRadArg(name, short, argUsage, description string, defaultValue []string) StringArrRslArg {
	return StringArrRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName:      name,
			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *StringArrRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.StringArrayVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *StringArrRslArg) SetValue(arg string) {
	// split on arg commas
	split := strings.Split(arg, ",")
	vals := make([]string, len(split))
	for i, v := range split {
		vals[i] = v
	}
	f.Value = vals
}

// --- int

type IntRslArg struct {
	BaseRslArg
	Value   int64
	Default int64
}

func NewIntRadArg(name, short, argUsage, description string, defaultValue int64) IntRslArg {
	return IntRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName: name,

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

func (f *IntRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Int64VarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *IntRslArg) SetValue(arg string) {
	parsed, err := strconv.Atoi(arg)
	if err != nil {
		RP.CtxErrorExit(NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected int, but could not parse: %v\n", arg)))
	}
	val := int64(parsed)
	f.Value = val
}

// --- int array

type IntArrRslArg struct {
	BaseRslArg
	Value   []int64
	Default []int64
}

func NewIntArrRadArg(name, short, argUsage, description string, defaultValue []int64) IntArrRslArg {
	return IntArrRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *IntArrRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Int64SliceVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *IntArrRslArg) SetValue(arg string) {
	// split on arg commas
	split := strings.Split(arg, ",")
	ints := make([]int64, len(split))
	for i, v := range split {
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			RP.CtxErrorExit(NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected int, but could not parse: %v\n", arg)))
		}
		ints[i] = parsed
	}
	f.Value = ints
}

// --- float

type FloatRslArg struct {
	BaseRslArg
	Value   float64
	Default float64
}

func NewFloatRadArg(name, short, argUsage, description string, defaultValue float64) FloatRslArg {
	return FloatRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName:      name,
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

func (f *FloatRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Float64VarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *FloatRslArg) SetValue(arg string) {
	parsed, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		RP.CtxErrorExit(NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected float, but could not parse: %v\n", arg)))
	}
	f.Value = parsed
}

// --- float array

type FloatArrRslArg struct {
	BaseRslArg
	Value   []float64
	Default []float64
}

func NewFloatArrRadArg(name, short, argUsage, description string, defaultValue []float64) FloatArrRslArg {
	return FloatArrRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          argUsage,
			Description:       description,
			defaultAsString:   ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault: len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *FloatArrRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Float64SliceVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *FloatArrRslArg) SetValue(arg string) {
	// split on arg commas
	split := strings.Split(arg, ",")
	floats := make([]float64, len(split))
	for i, v := range split {
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			RP.CtxErrorExit(NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected float, but could not parse: %v\n", arg)))
		}
		floats[i] = parsed
	}
	f.Value = floats
}

// --- MockResponse

type MockResponseRslArg struct {
	BaseRslArg
	Value MockResponseSlice
}

func NewMockResponseRadArg(name, short, usage string) MockResponseRslArg {
	return MockResponseRslArg{
		BaseRslArg: BaseRslArg{
			ExternalName: name,

			Identifier:        name,
			Short:             short,
			ArgUsage:          "string",
			Description:       usage,
			defaultAsString:   "",
			hasNonZeroDefault: false,
		},
	}
}

func (f *MockResponseRslArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.VarP(&f.Value, f.ExternalName, f.Short, f.Description)

	f.registered = true
}

func (f *MockResponseRslArg) SetValue(arg string) {
	RP.RadErrorExit(fmt.Sprintf("This function is expected to only be called for script args."+
		" MockResponse cannot be a script arg: %v\n", arg))
}

// --- general

func CreateFlag(arg *ScriptArg) RslArg {
	apiName, argType, shorthand, description := arg.ApiName, arg.Type, "", ""
	if arg.Short != nil {
		shorthand = *arg.Short
	}
	if arg.Description != nil {
		description = *arg.Description
	}

	switch argType {
	case ArgStringT:
		defVal := ""
		if arg.DefaultString != nil {
			defVal = *arg.DefaultString
		}
		f := NewStringRadArg(apiName, shorthand, "string", description, defVal, arg.EnumConstraint, arg.RegexConstraint)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgStringArrayT:
		var defVal []string
		if arg.DefaultStringList != nil {
			defVal = *arg.DefaultStringList
		}
		f := NewStringArrRadArg(apiName, shorthand, "string,string", description, defVal)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgIntT:
		defVal := int64(0)
		if arg.DefaultInt != nil {
			defVal = *arg.DefaultInt
		}
		f := NewIntRadArg(apiName, shorthand, "int", description, defVal)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgIntArrayT:
		var defVal []int64
		if arg.DefaultIntList != nil {
			defVal = *arg.DefaultIntList
		}
		f := NewIntArrRadArg(apiName, shorthand, "int,int", description, defVal)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgFloatT:
		defVal := 0.0
		if arg.DefaultFloat != nil {
			defVal = *arg.DefaultFloat
		}
		f := NewFloatRadArg(apiName, shorthand, "float", description, defVal)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgFloatArrayT:
		var defVal []float64
		if arg.DefaultFloatList != nil {
			defVal = *arg.DefaultFloatList
		}
		f := NewFloatArrRadArg(apiName, shorthand, "float,float", description, defVal)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgBoolT:
		defVal := false
		if arg.DefaultBool != nil {
			defVal = *arg.DefaultBool
		}
		f := NewBoolRadArg(apiName, shorthand, description, defVal)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgBoolArrayT:
		var defVal []bool
		if arg.DefaultBoolList != nil {
			defVal = *arg.DefaultBoolList
		}
		f := NewBoolArrRadArg(apiName, shorthand, "bool,bool", description, defVal)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	default:
		RP.RadNodeErrorExit(&arg.Decl, fmt.Sprintf("Unknown arg type: %v\n", argType))
		panic(UNREACHABLE)
	}
}
