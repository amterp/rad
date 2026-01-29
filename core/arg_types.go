package core

import (
	"fmt"
	"regexp"

	"github.com/amterp/rad/rts"
)

type ScriptArg struct {
	Name               string // Internal (as written in script)
	ExternalName       string // External (hyphenated for CLI)
	Decl               rts.ArgDecl
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
	decl rts.ArgDecl,
	enumConstraint *rts.ArgEnumConstraint,
	regexConstraint *rts.ArgRegexConstraint,
	rangeConstraint *rts.ArgRangeConstraint,
	requiresConstraint []string,
	excludesConstraint []string,
) *ScriptArg {
	name := decl.Name.Name
	externalName := decl.ExternalName()

	defaultVal := decl.Default

	scriptArg := &ScriptArg{
		Name:               name,
		ExternalName:       externalName,
		Decl:               decl,
		Short:              decl.ShorthandStr(),
		Type:               ToRadArgTypeT(decl.Type.Type),
		Description:        decl.CommentStr(),
		IsNullable:         decl.Optional != nil,
		HasDefaultValue:    hasDefaultValue(decl),
		IsVariadic:         decl.IsVariadic,
		EnumConstraint:     convertEnumConstraint(enumConstraint),
		RegexConstraint:    convertRegexConstraint(regexConstraint),
		RangeConstraint:    convertRangeConstraint(rangeConstraint),
		RequiresConstraint: requiresConstraint,
		ExcludesConstraint: excludesConstraint,
	}

	if defaultVal != nil {
		scriptArg.DefaultString = defaultVal.DefaultString
		scriptArg.DefaultInt = defaultVal.DefaultInt
		scriptArg.DefaultFloat = defaultVal.DefaultFloat
		scriptArg.DefaultBool = defaultVal.DefaultBool
		scriptArg.DefaultStringList = defaultVal.DefaultStringList
		scriptArg.DefaultIntList = defaultVal.DefaultIntList
		scriptArg.DefaultFloatList = defaultVal.DefaultFloatList
		scriptArg.DefaultBoolList = defaultVal.DefaultBoolList
	}

	return scriptArg
}

func convertEnumConstraint(constraint *rts.ArgEnumConstraint) *[]string {
	if constraint == nil {
		return nil
	}
	return &constraint.Values.Values
}

func convertRegexConstraint(constraint *rts.ArgRegexConstraint) *regexp.Regexp {
	if constraint == nil {
		return nil
	}
	regexStr := constraint.Regex.Value
	compiled, err := regexp.Compile(regexStr)
	if err != nil {
		RP.CtxErrorExit(NewCtxFromRtsNode(constraint, fmt.Sprintf("Invalid regex '%s': %s", regexStr, err.Error())))
	}
	return compiled
}

func convertRangeConstraint(constraint *rts.ArgRangeConstraint) *ArgRangeConstraint {
	if constraint == nil {
		return nil
	}

	rang := constraint.Range
	minInclusive := rang.Opener.Src() == "["
	maxInclusive := rang.Closer.Src() == "]"

	var maxV *float64
	if rang.Max != nil {
		maxV = &rang.Max.Value
	}

	var minV *float64
	if rang.Min != nil {
		minV = &rang.Min.Value
	}

	return &ArgRangeConstraint{
		Min:          minV,
		MinInclusive: minInclusive,
		Max:          maxV,
		MaxInclusive: maxInclusive,
	}
}

func hasDefaultValue(decl rts.ArgDecl) bool {
	if decl.Type.Type == "bool" {
		return true
	}
	return decl.Default != nil
}
