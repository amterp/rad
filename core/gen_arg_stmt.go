// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"regexp"
	"strings"
)

type ArgStmt interface {
	Accept(visitor ArgStmtVisitor)
}
type ArgStmtVisitor interface {
	VisitArgDeclarationArgStmt(ArgDeclaration)
	VisitArgEnumArgStmt(ArgEnum)
	VisitArgRegexArgStmt(ArgRegex)
}
type ArgDeclaration struct {
	Identifier Token
	Rename     *Token
	Flag       *Token
	ArgType    RslArgType
	IsOptional bool
	Default    *LiteralOrArray
	Comment    *ArgCommentToken
}

func (e ArgDeclaration) Accept(visitor ArgStmtVisitor) {
	visitor.VisitArgDeclarationArgStmt(e)
}
func (e ArgDeclaration) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Rename: %v", e.Rename))
	parts = append(parts, fmt.Sprintf("Flag: %v", e.Flag))
	parts = append(parts, fmt.Sprintf("ArgType: %v", e.ArgType))
	parts = append(parts, fmt.Sprintf("IsOptional: %v", e.IsOptional))
	parts = append(parts, fmt.Sprintf("Default: %v", e.Default))
	parts = append(parts, fmt.Sprintf("Comment: %v", e.Comment))
	return fmt.Sprintf("ArgDeclaration(%s)", strings.Join(parts, ", "))
}

type ArgEnum struct {
	EnumTkn    Token
	Identifier Token
	Values     MixedArrayLiteral
}

func (e ArgEnum) Accept(visitor ArgStmtVisitor) {
	visitor.VisitArgEnumArgStmt(e)
}
func (e ArgEnum) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("EnumTkn: %v", e.EnumTkn))
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("ArgEnum(%s)", strings.Join(parts, ", "))
}

type ArgRegex struct {
	RegexTkn   Token
	Identifier Token
	Regex      *regexp.Regexp
}

func (e ArgRegex) Accept(visitor ArgStmtVisitor) {
	visitor.VisitArgRegexArgStmt(e)
}
func (e ArgRegex) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("RegexTkn: %v", e.RegexTkn))
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Regex: %v", e.Regex))
	return fmt.Sprintf("ArgRegex(%s)", strings.Join(parts, ", "))
}
