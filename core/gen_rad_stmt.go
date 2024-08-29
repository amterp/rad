// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type RadStmt interface {
	Accept(visitor RadStmtVisitor)
}
type RadStmtVisitor interface {
	VisitFieldsRadStmt(Fields)
}
type Fields struct {
	Identifiers []Token
}

func (e Fields) Accept(visitor RadStmtVisitor) {
	visitor.VisitFieldsRadStmt(e)
}
func (e Fields) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifiers: %v", e.Identifiers))
	return fmt.Sprintf("Fields(%s)", strings.Join(parts, ", "))
}
