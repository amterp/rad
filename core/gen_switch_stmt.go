// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type SwitchStmt interface {
	Accept(visitor SwitchStmtVisitor)
}
type SwitchStmtVisitor interface {
	VisitSwitchCaseSwitchStmt(SwitchCase)
	VisitSwitchDefaultSwitchStmt(SwitchDefault)
}
type SwitchCase struct {
	CaseKeyword Token
	Keys        []StringLiteral
	Values      []Expr
}

func (e SwitchCase) Accept(visitor SwitchStmtVisitor) {
	visitor.VisitSwitchCaseSwitchStmt(e)
}
func (e SwitchCase) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("CaseKeyword: %v", e.CaseKeyword))
	parts = append(parts, fmt.Sprintf("Keys: %v", e.Keys))
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("SwitchCase(%s)", strings.Join(parts, ", "))
}

type SwitchDefault struct {
	DefaultKeyword Token
	Values         []Expr
}

func (e SwitchDefault) Accept(visitor SwitchStmtVisitor) {
	visitor.VisitSwitchDefaultSwitchStmt(e)
}
func (e SwitchDefault) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("DefaultKeyword: %v", e.DefaultKeyword))
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("SwitchDefault(%s)", strings.Join(parts, ", "))
}
