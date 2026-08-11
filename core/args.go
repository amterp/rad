package core

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	ra "github.com/amterp/ra"

	"github.com/amterp/rad/rts"
	"github.com/amterp/rad/rts/rl"

	"github.com/samber/lo"
)

// RegistrationMode defines how an argument should be registered on a Ra command.
// These modes enforce the correct combination of positional capability and Ra global behavior.
type RegistrationMode int

const (
	// AsScriptArg: Script args on root when NO commands exist
	// - Positional + flag capable (flagOnly=false)
	// - Not a Ra global (asRaGlobal=false)
	AsScriptArg RegistrationMode = iota

	// AsCommandArg: Command-specific args on a subcommand
	// - Positional + flag capable (flagOnly=false)
	// - Not a Ra global (asRaGlobal=false)
	AsCommandArg

	// AsScriptFlagOnly: Script args on subcommands when commands exist
	// - Flag-only (flagOnly=true)
	// - Not a Ra global (asRaGlobal=false)
	// - Script args are shared across commands but don't interfere with command positionals
	AsScriptFlagOnly

	// AsGlobalFlag: Rad's built-in global flags (--help, --version, --color, etc.)
	// - Flag-only (flagOnly=true)
	// - Ra global (asRaGlobal=true, inherits to all subcommands)
	AsGlobalFlag

	// AsSharedNamespaceArg: args declared on a namespace command, inherited by
	// every descendant.
	// - Flag-only (flagOnly=true) so they don't consume a sub-command's positional
	// - Ra global (asRaGlobal=true) so Ra cascades them, and their configured
	//   state, to descendants - which is what lets the flag appear either side
	//   of the sub-command on the path
	//
	// Riding Ra's globals is a means, not an end: Ra files globals under
	// "Global options" and hides them in short help, so a namespace's shared
	// args currently render in the wrong section of its own help. The fix is
	// upstream (render inherited flags as command args); once it lands, this
	// mode and AsScriptFlagOnly collapse into one - they are the same idea,
	// separated only because the root was never modeled as a tree node.
	AsSharedNamespaceArg
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
	GetDescription() string
	DefaultAsString() string
	HasNonZeroDefault() bool // todo
	GetType() RadArgTypeT
	Register(cmd *ra.Cmd, mode RegistrationMode)
	Configured() bool // configured by the user in some way
	IsDefined() bool  // either configured or has a default
	SetValue(value string)
	IsOptional() bool
	IsNullable() bool
	GetSpan() *rl.Span // nil if not a script arg
	Hidden(bool)
	IsHidden() bool
	Excludes(otherArg RadArg) bool
	IsVariadic() bool
}

type BaseRadArg struct {
	ExternalName       string // User-facing arg they'll set in CLI
	Identifier         string // Identifier in script. If non-script arg, then equal to ExternalName
	Short              string
	usagePlaceholder   string
	Description        string
	requiresConstraint []string // Identifiers, not external names
	excludesConstraint []string // Identifiers, not external names
	hasDefault         bool     // aka 'is optional'
	defaultAsString    string
	hasNonZeroDefault  bool
	registeredOn       map[*ra.Cmd]bool // Track which commands this arg is registered on
	manuallySet        bool
	scriptArg          *ScriptArg
	hidden             bool
	bypassValidation   bool // If true, this flag can bypass normal validation requirements
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
	return RRootCmd.Configured(f.ExternalName) || f.manuallySet
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

func (f *BaseRadArg) GetSpan() *rl.Span {
	if f.scriptArg == nil {
		return nil
	}
	span := f.scriptArg.Span
	return &span
}

// argErrorCtx creates an ErrorCtx from the arg declaration's span for error reporting.
func (f *BaseRadArg) argErrorCtx(msg string) ErrorCtx {
	if f.scriptArg == nil {
		return ErrorCtx{OneLiner: msg}
	}
	return NewCtxFromSpan(f.scriptArg.Src, f.scriptArg.Span, msg, "")
}

func (f *BaseRadArg) Hidden(hide bool) {
	f.hidden = hide
}

func (f *BaseRadArg) IsHidden() bool {
	return f.hidden
}

func (f *BaseRadArg) SetBypassValidation(bypass bool) {
	f.bypassValidation = bypass
}

// SetUsagePlaceholder overrides the value placeholder shown in usage, e.g.
// "line:value" for --reply. Rad's own flags are the only callers: a script's
// args take ra's placeholder for their type, which is what keeps the type
// names in --help consistent across every arg rad registers.
func (f *BaseRadArg) SetUsagePlaceholder(placeholder string) {
	f.usagePlaceholder = placeholder
}

func (f *BaseRadArg) Excludes(otherArg RadArg) bool {
	return lo.Contains(f.excludesConstraint, otherArg.GetIdentifier())
}

func (f *BaseRadArg) IsVariadic() bool {
	return f.scriptArg != nil && f.scriptArg.IsVariadic
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

func (f *BoolRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	var opts []ra.RegisterOption
	opts = append(opts, ra.WithGlobal(asRaGlobal))
	if f.bypassValidation {
		opts = append(opts, ra.WithBypassValidation(true))
	}

	err := ra.NewBool(f.ExternalName).
		SetShort(f.Short).
		SetDefault(f.Default).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetFlagOnly(flagOnly).
		SetCustomUsageType(f.usagePlaceholder).
		RegisterWithPtr(cmd, &f.Value, opts...)

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register bool arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
}

func (f *BoolRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	arg = strings.ToLower(arg)
	if arg == "true" || arg == "1" {
		f.Value = true
	} else if arg == "false" || arg == "0" {
		f.Value = false
	} else {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Expected bool, but could not parse: %v\n", arg)))
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

func (f *BoolListRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	arg := ra.NewBoolSlice(f.ExternalName).
		SetShort(f.Short).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetHiddenInShortHelp(asRaGlobal).
		SetFlagOnly(flagOnly).
		SetCustomUsageType(f.usagePlaceholder)

	if f.hasDefault {
		arg = arg.SetDefault(f.Default)
	}

	if f.IsNullable() {
		arg = arg.SetOptional(true)
	}

	if f.scriptArg != nil && f.scriptArg.IsVariadic {
		arg = arg.SetVariadic(true)
	}

	arg = applyLenConstraint(arg, f.scriptArg)

	err := arg.
		RegisterWithPtr(cmd, &f.Value, ra.WithGlobal(asRaGlobal))

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register bool list arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
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
			RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Expected bool, but could not parse: %v\n", arg)))
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

func (f *StringRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	arg := ra.NewString(f.ExternalName).
		SetShort(f.Short).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetCustomUsageType(f.usagePlaceholder).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetRegexConstraint(f.RegexConstraint).
		SetHiddenInShortHelp(asRaGlobal).
		SetFlagOnly(flagOnly)

	if f.hasDefault {
		arg = arg.SetDefault(f.Default)
	}

	if f.IsNullable() {
		arg = arg.SetOptional(true)
	}

	if f.EnumConstraint != nil {
		arg = arg.SetEnumConstraint(*f.EnumConstraint)
	}

	err := arg.RegisterWithPtr(cmd, &f.Value, ra.WithGlobal(asRaGlobal))

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register string arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
}

func (f *StringRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	f.Value = arg
}

func (f *StringRadArg) GetType() RadArgTypeT {
	return ArgStringT
}

// Per-element constraints reach ra from the ScriptArg rather than through the
// list constructors. Every CreateFlag branch already attaches the ScriptArg
// before Register runs, so reading it here means one forwarding site per
// element type instead of one per construction site - which is how these got
// dropped in the first place.
//
// Each helper forwards only what its element type can honor. ra would reject a
// mismatch, and rad's checker rejects it earlier still (RAD40024), so applying
// them blindly would just turn a clear diagnostic into a registration failure.

func applyStringElementConstraints(arg *ra.StringSliceFlag, sa *ScriptArg) *ra.StringSliceFlag {
	if sa == nil {
		return arg
	}
	if sa.EnumConstraint != nil {
		arg = arg.SetEnumConstraint(*sa.EnumConstraint)
	}
	if sa.RegexConstraint != nil {
		arg = arg.SetRegexConstraint(sa.RegexConstraint)
	}
	return arg
}

// applyLenConstraint bounds the number of values. Unlike the element
// constraints this suits every list element type, bools included.
func applyLenConstraint[T any](arg *ra.SliceFlag[T], sa *ScriptArg) *ra.SliceFlag[T] {
	if sa == nil || sa.LenConstraint == nil {
		return arg
	}
	lc := sa.LenConstraint
	if lc.Min != nil {
		arg = arg.SetMinLen(int(*lc.Min), lc.MinInclusive)
	}
	if lc.Max != nil {
		arg = arg.SetMaxLen(int(*lc.Max), lc.MaxInclusive)
	}
	return arg
}

func applyRangeElementConstraints[T int64 | float64](arg *ra.SliceFlag[T], sa *ScriptArg) *ra.SliceFlag[T] {
	if sa == nil || sa.RangeConstraint == nil {
		return arg
	}
	rc := sa.RangeConstraint
	if rc.Min != nil {
		arg = arg.SetMin(*rc.Min, rc.MinInclusive)
	}
	if rc.Max != nil {
		arg = arg.SetMax(*rc.Max, rc.MaxInclusive)
	}
	return arg
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

func (f *StringListRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	arg := ra.NewStringSlice(f.ExternalName).
		SetShort(f.Short).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetHiddenInShortHelp(asRaGlobal).
		SetFlagOnly(flagOnly).
		SetCustomUsageType(f.usagePlaceholder)

	if f.hasDefault {
		arg = arg.SetDefault(f.Default)
	}

	if f.IsNullable() {
		arg = arg.SetOptional(true)
	}

	if f.scriptArg != nil && f.scriptArg.IsVariadic {
		arg = arg.SetVariadic(true)
	}

	arg = applyStringElementConstraints(arg, f.scriptArg)
	arg = applyLenConstraint(arg, f.scriptArg)

	err := arg.
		RegisterWithPtr(cmd, &f.Value, ra.WithGlobal(asRaGlobal))

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register string list arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
}

func (f *StringListRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	// split on arg commas
	split := strings.Split(arg, ",")
	vals := make([]string, len(split))
	for i, v := range split {
		vals[i] = v
	}

	// For variadic arguments with list defaults, clear defaults on first user input
	if f.scriptArg != nil && f.scriptArg.IsVariadic && f.scriptArg.DefaultStringList != nil {
		// Check if current Value contains only the defaults
		defaults := *f.scriptArg.DefaultStringList
		if len(f.Value) == len(defaults) {
			allMatch := true
			for i, v := range f.Value {
				if i >= len(defaults) || v != defaults[i] {
					allMatch = false
					break
				}
			}
			if allMatch {
				// This is the first user input, replace defaults completely
				f.Value = vals
				return
			}
		}
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

func (f *IntRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	arg := ra.NewInt64(f.ExternalName).
		SetShort(f.Short).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetCustomUsageType("int").
		SetHiddenInShortHelp(asRaGlobal).
		SetFlagOnly(flagOnly)

	if f.hasDefault {
		arg = arg.SetDefault(f.Default)
	}

	if f.IsNullable() {
		arg = arg.SetOptional(true)
	}

	if f.RangeConstraint != nil {
		if f.RangeConstraint.Min != nil {
			arg = arg.SetMin(int64(*f.RangeConstraint.Min), (*f.RangeConstraint).MinInclusive)
		}
		if f.RangeConstraint.Max != nil {
			arg = arg.SetMax(int64(*f.RangeConstraint.Max), (*f.RangeConstraint).MaxInclusive)
		}
	}

	err := arg.
		RegisterWithPtr(cmd, &f.Value, ra.WithGlobal(asRaGlobal))

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register int arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
}

func (f *IntRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	parsed, err := strconv.Atoi(arg)
	if err != nil {
		RP.CtxErrorExit(
			f.argErrorCtx(fmt.Sprintf("Expected int, but could not parse: %v\n", arg)),
		)
	}
	val := int64(parsed)
	f.Value = val
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

func (f *IntListRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	arg := ra.NewInt64Slice(f.ExternalName).
		SetShort(f.Short).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetCustomUsageType("ints").
		SetHiddenInShortHelp(asRaGlobal).
		SetFlagOnly(flagOnly)

	if f.hasDefault {
		arg = arg.SetDefault(f.Default)
	}

	if f.IsNullable() {
		arg = arg.SetOptional(true)
	}

	if f.scriptArg != nil && f.scriptArg.IsVariadic {
		arg = arg.SetVariadic(true)
	}

	arg = applyRangeElementConstraints(arg, f.scriptArg)
	arg = applyLenConstraint(arg, f.scriptArg)

	err := arg.
		RegisterWithPtr(cmd, &f.Value, ra.WithGlobal(asRaGlobal))

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register int list arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
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
				f.argErrorCtx(fmt.Sprintf("Expected int, but could not parse: %v\n", arg)),
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

func (f *FloatRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	arg := ra.NewFloat64(f.ExternalName).
		SetShort(f.Short).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetHiddenInShortHelp(asRaGlobal).
		SetFlagOnly(flagOnly).
		SetCustomUsageType(f.usagePlaceholder)

	if f.hasDefault {
		arg = arg.SetDefault(f.Default)
	}

	if f.IsNullable() {
		arg = arg.SetOptional(true)
	}

	if f.RangeConstraint != nil {
		if f.RangeConstraint.Min != nil {
			arg = arg.SetMin(*f.RangeConstraint.Min, (*f.RangeConstraint).MinInclusive)
		}
		if f.RangeConstraint.Max != nil {
			arg = arg.SetMax(*f.RangeConstraint.Max, (*f.RangeConstraint).MaxInclusive)
		}
	}

	err := arg.
		RegisterWithPtr(cmd, &f.Value, ra.WithGlobal(asRaGlobal))

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register float arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
}

func (f *FloatRadArg) SetValue(arg string) {
	f.BaseRadArg.SetValue(arg)
	parsed, err := rts.ParseFloat(arg)
	if err != nil {
		RP.CtxErrorExit(
			f.argErrorCtx(fmt.Sprintf("Expected float, but could not parse: %v\n", arg)),
		)
	}
	f.Value = parsed
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

func (f *FloatListRadArg) Register(cmd *ra.Cmd, mode RegistrationMode) {
	if f.registeredOn == nil {
		f.registeredOn = make(map[*ra.Cmd]bool)
	}

	if f.registeredOn[cmd] {
		return
	}

	flagOnly, asRaGlobal := regModeToBoolFlags(mode)

	arg := ra.NewFloat64Slice(f.ExternalName).
		SetShort(f.Short).
		SetUsage(f.Description).
		SetHiddenInShortHelp(asRaGlobal).
		SetHidden(f.hidden).
		SetRequires(f.requiresConstraint).
		SetExcludes(f.excludesConstraint).
		SetHiddenInShortHelp(asRaGlobal).
		SetFlagOnly(flagOnly).
		SetCustomUsageType(f.usagePlaceholder)

	if f.hasDefault {
		arg = arg.SetDefault(f.Default)
	}

	if f.IsNullable() {
		arg = arg.SetOptional(true)
	}

	if f.scriptArg != nil && f.scriptArg.IsVariadic {
		arg = arg.SetVariadic(true)
	}

	arg = applyRangeElementConstraints(arg, f.scriptArg)
	arg = applyLenConstraint(arg, f.scriptArg)

	err := arg.
		RegisterWithPtr(cmd, &f.Value, ra.WithGlobal(asRaGlobal))

	if err != nil {
		RP.CtxErrorExit(f.argErrorCtx(fmt.Sprintf("Failed to register float list arg: %v\n", err)))
	}

	f.registeredOn[cmd] = true
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
				f.argErrorCtx(fmt.Sprintf("Expected float, but could not parse: %v\n", arg)),
			)
		}
		floats[i] = parsed
	}
	f.Value = floats
}

func (f *FloatListRadArg) GetType() RadArgTypeT {
	return ArgFloatListT
}

// --- general

func CreateFlag(arg *ScriptArg) RadArg {
	apiName, argType, shorthand, description := arg.ExternalName, arg.Type, "", ""
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

func (f *BaseRadArg) missingRequirement(required string) error {
	return fmt.Errorf("'%s' requires '%s', but '%s' was not set", f.ExternalName, required, required)
}

func (f *BaseRadArg) excludesRequirement(excluded string) error {
	return fmt.Errorf("'%s' excludes '%s', but '%s' was set", f.ExternalName, excluded, excluded)
}

func regModeToBoolFlags(mode RegistrationMode) (bool, bool) {
	var flagOnly, asRaGlobal bool
	switch mode {
	case AsScriptArg, AsCommandArg:
		flagOnly = false
		asRaGlobal = false
	case AsScriptFlagOnly:
		flagOnly = true
		asRaGlobal = false
	case AsGlobalFlag, AsSharedNamespaceArg:
		flagOnly = true
		asRaGlobal = true
	default:
		panic(fmt.Sprintf("Unknown RegistrationMode: %v", mode))
	}
	return flagOnly, asRaGlobal
}
