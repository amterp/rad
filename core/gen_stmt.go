// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type Stmt interface {
	Accept(visitor StmtVisitor)
}
type StmtVisitor interface {
	VisitExpressionStmt(*Expression)
	VisitPrimaryAssignStmt(*PrimaryAssign)
	VisitEmptyStmt(*Empty)
}
type Expression struct {
	expression Expr
}

func (e *Expression) Accept(visitor StmtVisitor) {
	visitor.VisitExpressionStmt(e)
}
func (e *Expression) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("expression: %v", e.expression))
	return fmt.Sprintf("Expression(%s)", strings.Join(parts, ", "))
}

type PrimaryAssign struct {
	name        Token
	initializer Expr
}

func (e *PrimaryAssign) Accept(visitor StmtVisitor) {
	visitor.VisitPrimaryAssignStmt(e)
}
func (e *PrimaryAssign) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("name: %v", e.name))
	parts = append(parts, fmt.Sprintf("initializer: %v", e.initializer))
	return fmt.Sprintf("PrimaryAssign(%s)", strings.Join(parts, ", "))
}

type Empty struct {
}

func (e *Empty) Accept(visitor StmtVisitor) {
	visitor.VisitEmptyStmt(e)
}
func (e *Empty) String() string {
	return "Empty()"
}
