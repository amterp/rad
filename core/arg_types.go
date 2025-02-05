package core

import (
	"fmt"
	"regexp"

	"github.com/amterp/rts"
)

type ScriptArg struct {
	Name            string // identifier name in the script
	ApiName         string // name that the user will see
	TreeNode        rts.ArgDecl
	Short           *string
	Type            RslArgTypeT
	Description     *string
	IsOptional      bool
	EnumConstraint  *[]string
	RegexConstraint *regexp.Regexp
	// first check the Type and IsOptional, then get the value
	DefaultString     *string
	DefaultStringList *[]string
	DefaultInt        *int64
	DefaultIntList    *[]int64
	DefaultFloat      *float64
	DefaultFloatList  *[]float64
	DefaultBool       *bool
	DefaultBoolList   *[]bool
}

func FromArgDecl(decl rts.ArgDecl, enumConstraint *rts.ArgEnumConstraint, regexConstraint *rts.ArgRegexConstraint) *ScriptArg {
	name := decl.Name.Name
	externalName := decl.ExternalName()

	defaultVal := decl.Default
	scriptArg := &ScriptArg{
		Name:            name,
		ApiName:         externalName,
		TreeNode:        decl,
		Short:           decl.ShorthandStr(),
		Type:            ToRslArgTypeT(decl.Type.Type),
		Description:     decl.CommentStr(),
		IsOptional:      isOptional(decl),
		EnumConstraint:  convertEnumConstraint(enumConstraint),
		RegexConstraint: convertRegexConstraint(regexConstraint),
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

func isOptional(decl rts.ArgDecl) bool {
	if decl.Type.Type == "bool" {
		return true
	}
	return decl.Default != nil
}
