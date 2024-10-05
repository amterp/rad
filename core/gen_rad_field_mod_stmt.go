// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type RadFieldModStmt interface {
	Accept(visitor RadFieldModStmtVisitor)
}
type RadFieldModStmtVisitor interface {
	VisitTruncateRadFieldModStmt(Truncate)
}
type Truncate struct {
	TruncToken Token
	Value      Expr
}

func (e Truncate) Accept(visitor RadFieldModStmtVisitor) {
	visitor.VisitTruncateRadFieldModStmt(e)
}
func (e Truncate) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("TruncToken: %v", e.TruncToken))
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("Truncate(%s)", strings.Join(parts, ", "))
}
