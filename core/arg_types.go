package core

import (
	"fmt"
	"regexp"

	"github.com/amterp/rad/rts/rl"
)

type ScriptArg struct {
	Name               string // Internal (as written in script)
	ExternalName       string // External (hyphenated for CLI)
	Span               rl.Span
	Src                string // Full script source, for error context rendering
	Short              *string
	Type               RadArgTypeT
	Description        *string
	IsNullable         bool // aka is optional. e.g. 'string?' syntax
	HasDefaultValue    bool
	IsVariadic         bool // whether this is a variadic argument (*options str)
	EnumConstraint     *[]string
	RegexConstraint    *regexp.Regexp
	RangeConstraint    *ArgRangeConstraint
	RequiresConstraint []string
	ExcludesConstraint []string
	// first check the Type and HasDefaultValue, then get the value
	DefaultString     *string
	DefaultStringList *[]string
	DefaultInt        *int64
	DefaultIntList    *[]int64
	DefaultFloat      *float64
	DefaultFloatList  *[]float64
	DefaultBool       *bool
	DefaultBoolList   *[]bool
}

type ArgRangeConstraint struct {
	Min          *float64
	MinInclusive bool
	Max          *float64
	MaxInclusive bool
}

func FromArgDecl(
	decl rl.ArgDecl,
	src string,
	enumConstraint *rl.ArgEnumConstraint,
	regexConstraint *rl.ArgRegexConstraint,
	rangeConstraint *rl.ArgRangeConstraint,
	requiresConstraint []string,
	excludesConstraint []string,
) *ScriptArg {
	name := decl.Name
	externalName := decl.ExternalName()

	scriptArg := &ScriptArg{
		Name:               name,
		ExternalName:       externalName,
		Span:               decl.Span(),
		Src:                src,
		Short:              decl.Shorthand,
		Type:               ToRadArgTypeT(decl.TypeName),
		Description:        decl.Comment,
		IsNullable:         decl.IsOptional,
		HasDefaultValue:    hasDefaultValue(decl),
		IsVariadic:         decl.IsVariadic,
		EnumConstraint:     convertEnumConstraint(enumConstraint),
		RegexConstraint:    convertRegexConstraint(regexConstraint, src),
		RangeConstraint:    convertRangeConstraint(rangeConstraint),
		RequiresConstraint: requiresConstraint,
		ExcludesConstraint: excludesConstraint,
	}

	scriptArg.DefaultString = decl.DefaultString
	scriptArg.DefaultInt = decl.DefaultInt
	scriptArg.DefaultFloat = decl.DefaultFloat
	scriptArg.DefaultBool = decl.DefaultBool
	scriptArg.DefaultStringList = decl.DefaultStringList
	scriptArg.DefaultIntList = decl.DefaultIntList
	scriptArg.DefaultFloatList = decl.DefaultFloatList
	scriptArg.DefaultBoolList = decl.DefaultBoolList

	return scriptArg
}

func convertEnumConstraint(constraint *rl.ArgEnumConstraint) *[]string {
	if constraint == nil {
		return nil
	}
	return &constraint.Values
}

func convertRegexConstraint(constraint *rl.ArgRegexConstraint, src string) *regexp.Regexp {
	if constraint == nil {
		return nil
	}
	regexStr := constraint.Value
	compiled, err := regexp.Compile(regexStr)
	if err != nil {
		RP.CtxErrorExit(NewCtxFromSpan(src, constraint.Span_, fmt.Sprintf("Invalid regex '%s': %s", regexStr, err.Error()), ""))
	}
	return compiled
}

func convertRangeConstraint(constraint *rl.ArgRangeConstraint) *ArgRangeConstraint {
	if constraint == nil {
		return nil
	}

	minInclusive := constraint.OpenerToken == "["
	maxInclusive := constraint.CloserToken == "]"

	return &ArgRangeConstraint{
		Min:          constraint.Min,
		MinInclusive: minInclusive,
		Max:          constraint.Max,
		MaxInclusive: maxInclusive,
	}
}

func hasDefaultValue(decl rl.ArgDecl) bool {
	if decl.TypeName == "bool" {
		return true
	}
	return decl.Default != nil
}
