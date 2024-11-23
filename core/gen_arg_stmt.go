// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type ArgStmt interface {
	Accept(visitor ArgStmtVisitor)
}
type ArgStmtVisitor interface {
	VisitArgDeclarationArgStmt(ArgDeclaration)
	VisitArgRestDeclArgStmt(ArgRestDecl)
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

type ArgRestDecl struct {
	Identifier Token
	Comment    *ArgCommentToken
}

func (e ArgRestDecl) Accept(visitor ArgStmtVisitor) {
	visitor.VisitArgRestDeclArgStmt(e)
}
func (e ArgRestDecl) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Comment: %v", e.Comment))
	return fmt.Sprintf("ArgRestDecl(%s)", strings.Join(parts, ", "))
}
