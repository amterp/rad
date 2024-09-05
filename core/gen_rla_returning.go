// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type RlaReturning interface {
	Accept(visitor RlaReturningVisitor) []RuntimeLiteral
}
type RlaReturningVisitor interface {
	VisitSwitchBlockRlaReturning(SwitchBlock) []RuntimeLiteral
}
type SwitchBlock struct {
	SwitchToken   Token
	Discriminator *Token
	Stmts         []SwitchStmt
}

func (e SwitchBlock) Accept(visitor RlaReturningVisitor) []RuntimeLiteral {
	return visitor.VisitSwitchBlockRlaReturning(e)
}
func (e SwitchBlock) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("SwitchToken: %v", e.SwitchToken))
	parts = append(parts, fmt.Sprintf("Discriminator: %v", e.Discriminator))
	parts = append(parts, fmt.Sprintf("Stmts: %v", e.Stmts))
	return fmt.Sprintf("SwitchBlock(%s)", strings.Join(parts, ", "))
}
