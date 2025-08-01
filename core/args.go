package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/amterp/rad/rts/rl"

	"github.com/amterp/rad/rts"

	ts "github.com/tree-sitter/go-tree-sitter"

	"github.com/samber/lo"
)

type ConstraintCtx struct {
	ScriptArgs map[string]RadArg // Identifier -> RadArg
}

func NewConstraintCtx(scriptArgs []RadArg) ConstraintCtx {
	scriptArgByIdentifier := make(map[string]RadArg)
	for _, arg := range scriptArgs {
		scriptArgByIdentifier[arg.GetIdentifier()] = arg
	}

	return ConstraintCtx{
		ScriptArgs: scriptArgByIdentifier,
	}
}

type RadArg interface {
	GetExternalName() string
	GetIdentifier() string
	GetShort() string
	GetArgUsage() string
	GetDescription() string
	DefaultAsString() string
	HasNonZeroDefault() bool // todo
	GetType() RadArgTypeT
	Register()
	Configured() bool // configured by the user in some way
	IsDefined() bool  // either configured or has a default
	SetValue(value string)
	IsOptional() bool
	IsNullable() bool
	GetNode() *ts.Node // nil if not a script arg
	Hidden(bool)
	IsHidden() bool
	// isolated constraints on one arg against its given value
	ValidateConstraints() error
	// constraints between arguments
	ValidateRelationalConstraints(ConstraintCtx) error
	Excludes(otherArg RadArg) bool
}

type BaseRadArg struct {
	ExternalName       string // User-facing arg they'll set in CLI
	Identifier         string // Identifier in script. If non-script arg, then equal to ExternalName
	Short              string
	ArgUsage           string
	Description        string
	requiresConstraint []string // Identifiers, not external names
	excludesConstraint []string // Identifiers, not external names
	hasDefault         bool     // aka 'is optional'
	defaultAsString    string
	hasNonZeroDefault  bool
	registered         bool
	manuallySet        bool
	scriptArg          *ScriptArg
	hidden             bool
}

func (f *BaseRadArg) GetExternalName() string {
	return f.ExternalName
}

func (f *BaseRadArg) GetIdentifier() string {
	return f.Identifier
}

func (f *BaseRadArg) GetShort() string {
	return f.Short
}

func (f *BaseRadArg) GetArgUsage() string {
	return f.ArgUsage
}

func (f *BaseRadArg) GetDescription() string {
	return f.Description
}

func (f *BaseRadArg) DefaultAsString() string {
	return f.defaultAsString
}

func (f *BaseRadArg) HasNonZeroDefault() bool {
	return f.hasNonZeroDefault
}

func (f *BaseRadArg) Configured() bool {
	return RFlagSet.Lookup(f.ExternalName).Changed || f.manuallySet
}

func (f *BaseRadArg) IsDefined() bool {
	return f.Configured() || f.hasDefault
}

func (f *BaseRadArg) SetValue(_ string) {
	f.manuallySet = true
}

func (f *BaseRadArg) IsOptional() bool {
	if f.scriptArg == nil {
		// global args are indeed optional
		return true
	}

	return f.scriptArg.HasDefaultValue || f.scriptArg.IsNullable
}

func (f *BaseRadArg) IsNullable() bool {
	if f.scriptArg == nil {
		return false
	}

	return f.scriptArg.IsNullable
}

func (f *BaseRadArg) GetNode() *ts.Node {
	if f.scriptArg == nil {
		return nil
	}

	return f.scriptArg.Decl.Node()
}

func (f *BaseRadArg) Hidden(hide bool) {
	f.hidden = hide
}

func (f *BaseRadArg) IsHidden() bool {
	return f.hidden
}

func (f *BaseRadArg) ValidateConstraints() error {
	// Base impl does nothing -- each arg type will implement its own constraints
	return nil
}

func (f *BaseRadArg) ValidateRelationalConstraints(ctx ConstraintCtx) error {
	requires := f.requiresConstraint

	if !f.IsDefined() {
		// relational constraints only apply to defined args
		return nil
	}

	thisArg := ctx.ScriptArgs[f.Identifier]
	if thisBoolArg, ok := thisArg.(*BoolRadArg); ok {
		if !thisBoolArg.Value {
			// this bool arg is false, so its constraints are not relevant
			return nil
		}
	}

	for _, required := range requires {
		reqArg := ctx.ScriptArgs[required]
		required = reqArg.GetExternalName()

		if boolArg, ok := reqArg.(*BoolRadArg); ok {
			if !boolArg.Value {
				// bool arg is false but is required
				return f.missingRequirement(required)
			}
			return nil
		}

		if !reqArg.IsDefined() {
			return f.missingRequirement(required)
		}
	}

	for _, excluded := range f.excludesConstraint {
		exclArg := ctx.ScriptArgs[excluded]
		excluded = exclArg.GetExternalName()

		if boolArg, ok := exclArg.(*BoolRadArg); ok {
			if boolArg.Value {
				// bool arg is true but is excluded
				return f.excludesRequirement(excluded)
			}
			return nil
		}

		if exclArg.IsDefined() {
			return f.excludesRequirement(excluded)
		}
	}

	return nil
}

func (f *BaseRadArg) Excludes(otherArg RadArg) bool {
	return lo.Contains(f.excludesConstraint, otherArg.GetIdentifier())
}

// --- bool

type BoolRadArg struct {
	BaseRadArg
	Value   bool
	Default bool
}

func NewBoolRadArg(name,
	short,
	description string,
	hasDefault bool,
	defaultValue bool,
	requires,
	excludes []string,
) BoolRadArg {
	return BoolRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           "",
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    fmt.Sprint(defaultValue),
			hasNonZeroDefault:  defaultValue != false,
		},
		Default: defaultValue,
	}
}

func (f *BoolRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.BoolVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *BoolRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	arg = strings.ToLower(arg)
	if arg == "true" || arg == "1" {
		f.Value = true
	} else if arg == "false" || arg == "0" {
		f.Value = false
	} else {
		RP.CtxErrorExit(NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected bool, but could not parse: %v\n", arg)))
	}
}

func (f *BoolRadArg) GetType() RadArgTypeT {
	return ArgBoolT
}

// --- bool array

type BoolListRadArg struct {
	BaseRadArg
	Value   []bool
	Default []bool
}

func NewBoolListRadArg(name,
	short,
	argUsage,
	description string,
	hasDefault bool,
	defaultValue []bool,
	requires,
	excludes []string,
) BoolListRadArg {
	return BoolListRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           argUsage,
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault:  len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *BoolListRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.BoolSliceVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *BoolListRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
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

func (f *BoolListRadArg) GetType() RadArgTypeT {
	return ArgBoolListT
}

// --- string

type StringRadArg struct {
	BaseRadArg
	Value           string
	Default         string
	EnumConstraint  *[]string
	RegexConstraint *regexp.Regexp
}

func NewStringRadArg(
	name,
	short,
	argUsage,
	description string,
	hasDefault bool,
	defaultValue string,
	enum *[]string,
	regex *regexp.Regexp,
	requires,
	excludes []string,
) StringRadArg {
	return StringRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           argUsage,
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    defaultValue,
			hasNonZeroDefault:  defaultValue != "",
		},
		Default:         defaultValue,
		EnumConstraint:  enum,
		RegexConstraint: regex,
	}
}

func (f *StringRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.StringVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *StringRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	f.Value = arg
}

func (f *StringRadArg) GetDescription() string {
	var sb strings.Builder

	sb.WriteString(f.Description)

	if f.EnumConstraint != nil {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString("Valid values: [")
		sb.WriteString(strings.Join(*f.EnumConstraint, ", "))
		sb.WriteString("].")
	}

	if f.RegexConstraint != nil {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString("Regex: ")
		sb.WriteString((*f.RegexConstraint).String())
	}

	return sb.String()
}

//goland:noinspection GoErrorStringFormat
func (f *StringRadArg) ValidateConstraints() error {
	err := f.BaseRadArg.ValidateConstraints()
	if err != nil {
		return err
	}

	if f.EnumConstraint != nil {
		if !lo.Contains(*f.EnumConstraint, f.Value) {
			return fmt.Errorf(
				"Invalid '%s' value: %v (valid values: %s)",
				f.ExternalName,
				f.Value,
				strings.Join(*f.EnumConstraint, ", "),
			)
		}
	}

	constraint := f.RegexConstraint
	if constraint != nil {
		if !constraint.MatchString(f.Value) {
			return fmt.Errorf(
				"Invalid '%s' value: %v (must match regex: %s)",
				f.ExternalName,
				f.Value,
				constraint.String(),
			)
		}
	}

	return nil
}

func (f *StringRadArg) GetType() RadArgTypeT {
	return ArgStringT
}

// --- string array

type StringListRadArg struct {
	BaseRadArg
	Value   []string
	Default []string
}

func NewStringListRadArg(
	name,
	short,
	argUsage,
	description string,
	hasDefault bool,
	defaultValue,
	requires,
	excludes []string,
) StringListRadArg {
	return StringListRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           argUsage,
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault:  len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *StringListRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.StringArrayVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *StringListRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	// split on arg commas
	split := strings.Split(arg, ",")
	vals := make([]string, len(split))
	for i, v := range split {
		vals[i] = v
	}
	f.Value = vals
}

func (f *StringListRadArg) GetType() RadArgTypeT {
	return ArgStrListT
}

// --- int

type IntRadArg struct {
	BaseRadArg
	Value           int64
	Default         int64
	RangeConstraint *ArgRangeConstraint
}

func NewIntRadArg(
	name,
	short,
	argUsage,
	description string,
	hasDefault bool,
	defaultValue int64,
	rangeConstraint *ArgRangeConstraint,
	requires,
	excludes []string,
) IntRadArg {
	return IntRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           argUsage,
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    ToPrintable(defaultValue),
			hasNonZeroDefault:  defaultValue != 0,
		},
		Default:         defaultValue,
		RangeConstraint: rangeConstraint,
	}
}

func (f *IntRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Int64VarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *IntRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	parsed, err := strconv.Atoi(arg)
	if err != nil {
		RP.CtxErrorExit(
			NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected int, but could not parse: %v\n", arg)),
		)
	}
	val := int64(parsed)
	f.Value = val
}

func (f *IntRadArg) GetDescription() string {
	var sb strings.Builder

	sb.WriteString(f.Description)
	addRangeDescriptionIfPresent(&sb, f.RangeConstraint)

	return sb.String()
}

func (f *IntRadArg) ValidateConstraints() error {
	err := f.BaseRadArg.ValidateConstraints()
	if err != nil {
		return err
	}
	validateRange(f.scriptArg.Decl, float64(f.Value), f.RangeConstraint)
	return nil // todo validateRange should return error
}

func (f *IntRadArg) GetType() RadArgTypeT {
	return ArgIntT
}

// --- int array

type IntListRadArg struct {
	BaseRadArg
	Value   []int64
	Default []int64
}

func NewIntListRadArg(
	name,
	short,
	argUsage,
	description string,
	hasDefault bool,
	defaultValue []int64,
	requires,
	excludes []string,
) IntListRadArg {
	return IntListRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           argUsage,
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault:  len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *IntListRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Int64SliceVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *IntListRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	// split on arg commas
	split := strings.Split(arg, ",")
	ints := make([]int64, len(split))
	for i, v := range split {
		parsed, err := rts.ParseInt(v)
		if err != nil {
			RP.CtxErrorExit(
				NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected int, but could not parse: %v\n", arg)),
			)
		}
		ints[i] = parsed
	}
	f.Value = ints
}

func (f *IntListRadArg) GetType() RadArgTypeT {
	return ArgIntListT
}

// --- float

type FloatRadArg struct {
	BaseRadArg
	Value           float64
	Default         float64
	RangeConstraint *ArgRangeConstraint
}

func NewFloatRadArg(
	name,
	short,
	argUsage,
	description string,
	hasDefault bool,
	defaultValue float64,
	constraint *ArgRangeConstraint,
	requires,
	excludes []string,
) FloatRadArg {
	return FloatRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           argUsage,
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    ToPrintable(defaultValue),
			hasNonZeroDefault:  defaultValue != 0,
		},
		Default:         defaultValue,
		RangeConstraint: constraint,
	}
}

func (f *FloatRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Float64VarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *FloatRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	parsed, err := rts.ParseFloat(arg)
	if err != nil {
		RP.CtxErrorExit(
			NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected float, but could not parse: %v\n", arg)),
		)
	}
	f.Value = parsed
}

func (f *FloatRadArg) GetDescription() string {
	var sb strings.Builder

	sb.WriteString(f.Description)
	addRangeDescriptionIfPresent(&sb, f.RangeConstraint)

	return sb.String()
}

func (f *FloatRadArg) ValidateConstraints() error {
	err := f.BaseRadArg.ValidateConstraints()
	if err != nil {
		return err
	}
	validateRange(f.scriptArg.Decl, f.Value, f.RangeConstraint)
	return nil // todo validateRange should return error
}

func (f *FloatRadArg) GetType() RadArgTypeT {
	return ArgFloatT
}

// --- float array

type FloatListRadArg struct {
	BaseRadArg
	Value   []float64
	Default []float64
}

func NewFloatListRadArg(
	name,
	short,
	argUsage,
	description string,
	hasDefault bool,
	defaultValue []float64,
	requires,
	excludes []string,
) FloatListRadArg {
	return FloatListRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:       name,
			Identifier:         name,
			Short:              short,
			ArgUsage:           argUsage,
			Description:        description,
			requiresConstraint: requires,
			excludesConstraint: excludes,
			hasDefault:         hasDefault,
			defaultAsString:    ToPrintable(convertToInterfaceArr(defaultValue)),
			hasNonZeroDefault:  len(defaultValue) > 0,
		},
		Default: defaultValue,
	}
}

func (f *FloatListRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.Float64SliceVarP(&f.Value, f.ExternalName, f.Short, f.Default, f.Description)

	f.registered = true
}

func (f *FloatListRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	// split on arg commas
	split := strings.Split(arg, ",")
	floats := make([]float64, len(split))
	for i, v := range split {
		parsed, err := rts.ParseFloat(v)
		if err != nil {
			RP.CtxErrorExit(
				NewCtxFromRtsNode(&f.scriptArg.Decl, fmt.Sprintf("Expected float, but could not parse: %v\n", arg)),
			)
		}
		floats[i] = parsed
	}
	f.Value = floats
}

func (f *FloatListRadArg) GetType() RadArgTypeT {
	return ArgFloatListT
}

// --- MockResponse

type MockResponseRadArg struct {
	BaseRadArg
	Value MockResponseSlice
}

func NewMockResponseRadArg(name, short, usage string) MockResponseRadArg {
	return MockResponseRadArg{
		BaseRadArg: BaseRadArg{
			ExternalName:      name,
			Identifier:        name,
			Short:             short,
			ArgUsage:          rl.T_STR,
			Description:       usage,
			defaultAsString:   "",
			hasNonZeroDefault: false,
		},
	}
}

func (f *MockResponseRadArg) Register() {
	if f.registered {
		return
	}

	RFlagSet.VarP(&f.Value, f.ExternalName, f.Short, f.Description)

	f.registered = true
}

func (f *MockResponseRadArg) SetValue(arg string) {
	RP.RadErrorExit(fmt.Sprintf("This function is expected to only be called for script args."+
		" MockResponse cannot be a script arg: %v\n", arg))
}

func (f *MockResponseRadArg) GetType() RadArgTypeT {
	return ArgStringT // hmm not really
}

// --- general

func CreateFlag(arg *ScriptArg) RadArg {
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
		hasDefault := arg.DefaultString != nil
		if hasDefault {
			defVal = *arg.DefaultString
		}
		f := NewStringRadArg(
			apiName,
			shorthand,
			rl.T_STR,
			description,
			hasDefault,
			defVal,
			arg.EnumConstraint,
			arg.RegexConstraint,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgStrListT:
		var defVal []string
		hasDefault := arg.DefaultStringList != nil
		if hasDefault {
			defVal = *arg.DefaultStringList
		}
		f := NewStringListRadArg(
			apiName,
			shorthand,
			"string,string",
			description,
			hasDefault,
			defVal,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgIntT:
		defVal := int64(0)
		hasDefault := arg.DefaultInt != nil
		if hasDefault {
			defVal = *arg.DefaultInt
		}
		f := NewIntRadArg(
			apiName,
			shorthand,
			"int",
			description,
			hasDefault,
			defVal,
			arg.RangeConstraint,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgIntListT:
		var defVal []int64
		hasDefault := arg.DefaultIntList != nil
		if hasDefault {
			defVal = *arg.DefaultIntList
		}
		f := NewIntListRadArg(
			apiName,
			shorthand,
			"int,int",
			description,
			hasDefault,
			defVal,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgFloatT:
		defVal := 0.0
		hasDefault := arg.DefaultFloat != nil
		if hasDefault {
			defVal = *arg.DefaultFloat
		}
		f := NewFloatRadArg(
			apiName,
			shorthand,
			"float",
			description,
			hasDefault,
			defVal,
			arg.RangeConstraint,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgFloatListT:
		var defVal []float64
		hasDefault := arg.DefaultFloatList != nil
		if hasDefault {
			defVal = *arg.DefaultFloatList
		}
		f := NewFloatListRadArg(
			apiName,
			shorthand,
			"float,float",
			description,
			hasDefault,
			defVal,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgBoolT:
		defVal := false
		if arg.DefaultBool != nil {
			defVal = *arg.DefaultBool
		}
		f := NewBoolRadArg(
			apiName,
			shorthand,
			description,
			true,
			defVal,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	case ArgBoolListT:
		var defVal []bool
		hasDefault := arg.DefaultBoolList != nil
		if hasDefault {
			defVal = *arg.DefaultBoolList
		}
		f := NewBoolListRadArg(
			apiName,
			shorthand,
			"bool,bool",
			description,
			hasDefault,
			defVal,
			arg.RequiresConstraint,
			arg.ExcludesConstraint,
		)
		f.scriptArg = arg
		f.Identifier = arg.Name
		return &f
	default:
		panic(fmt.Sprintf("Unhandled arg type: %v", argType))
	}
}

func convertToInterfaceArr[T any](i []T) []interface{} {
	converted := make([]interface{}, len(i))
	for j, v := range i {
		converted[j] = v
	}
	return converted
}

func validateRange(decl rts.ArgDecl, val float64, rangeConstraint *ArgRangeConstraint) {
	if rangeConstraint == nil {
		return
	}

	rMin := rangeConstraint.Min
	if rMin != nil {
		if rangeConstraint.MinInclusive {
			if val < *rMin {
				RP.CtxErrorExit(
					NewCtxFromRtsNode(
						&decl,
						fmt.Sprintf("'%s' value %v is < minimum %v", decl.ExternalName(), val, *rMin),
					),
				)
			}
		} else {
			if val <= *rMin {
				RP.CtxErrorExit(NewCtxFromRtsNode(&decl, fmt.Sprintf("'%s' value %v is <= minimum (exclusive) %v", decl.ExternalName(), val, *rMin)))
			}
		}
	}

	rMax := rangeConstraint.Max
	if rMax != nil {
		if rangeConstraint.MaxInclusive {
			if val > *rMax {
				RP.CtxErrorExit(
					NewCtxFromRtsNode(
						&decl,
						fmt.Sprintf("'%s' value %v is > maximum %v", decl.ExternalName(), val, *rMax),
					),
				)
			}
		} else {
			if val >= *rMax {
				RP.CtxErrorExit(NewCtxFromRtsNode(&decl, fmt.Sprintf("'%s' value %v is >= maximum (exclusive) %v", decl.ExternalName(), val, *rMax)))
			}
		}
	}
}

func addRangeDescriptionIfPresent(sb *strings.Builder, rangeConstraint *ArgRangeConstraint) {
	if rangeConstraint != nil {
		if sb.Len() > 0 {
			sb.WriteString(" ")
		}
		sb.WriteString("Range: ")
		sb.WriteString(lo.Ternary(rangeConstraint.MinInclusive, "[", "("))
		if rangeConstraint.Min != nil {
			sb.WriteString(fmt.Sprintf("%v", *rangeConstraint.Min))
		}
		sb.WriteString(", ")
		if rangeConstraint.Max != nil {
			sb.WriteString(fmt.Sprintf("%v", *rangeConstraint.Max))
		}
		sb.WriteString(lo.Ternary(rangeConstraint.MaxInclusive, "]", ")"))
	}
}

func (f *BaseRadArg) missingRequirement(required string) error {
	return fmt.Errorf("'%s' requires '%s', but '%s' was not set", f.ExternalName, required, required)
}

func (f *BaseRadArg) excludesRequirement(excluded string) error {
	return fmt.Errorf("'%s' excludes '%s', but '%s' was set", f.ExternalName, excluded, excluded)
}
