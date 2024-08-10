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
	VisitArgDeclarationArgStmt(*ArgDeclaration)
}
type ArgDeclaration struct {
	identifier  Token
	rename      *Token
	flag        *Token
	argType     RslType
	isOptional  bool
	defaultInit *Expr
	comment     ArgCommentToken
}

func (e *ArgDeclaration) Accept(visitor ArgStmtVisitor) {
	visitor.VisitArgDeclarationArgStmt(e)
}
func (e *ArgDeclaration) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("identifier: %v", e.identifier))
	parts = append(parts, fmt.Sprintf("rename: %v", e.rename))
	parts = append(parts, fmt.Sprintf("flag: %v", e.flag))
	parts = append(parts, fmt.Sprintf("argType: %v", e.argType))
	parts = append(parts, fmt.Sprintf("isOptional: %v", e.isOptional))
	parts = append(parts, fmt.Sprintf("defaultInit: %v", e.defaultInit))
	parts = append(parts, fmt.Sprintf("comment: %v", e.comment))
	return fmt.Sprintf("ArgDeclaration(%s)", strings.Join(parts, ", "))
}
