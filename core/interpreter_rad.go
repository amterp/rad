package core

import (
	"fmt"
)

type RadBlockInterpreter struct {
	i          *MainInterpreter
	invocation *radInvocationOld
}

func NewRadBlockInterpreter(i *MainInterpreter) *RadBlockInterpreter {
	return &RadBlockInterpreter{i: i}
}

func (r RadBlockInterpreter) Run(block RadBlock) {
	// DELETE
}

func (r RadBlockInterpreter) VisitFieldsRadStmt(fields Fields) {
	// DELETE
}

func (r RadBlockInterpreter) VisitFieldModsRadStmt(mods FieldMods) {
	// DELETE
}

func (r RadBlockInterpreter) VisitSortRadStmt(sort Sort) {
	// DELETE
}

func (r RadBlockInterpreter) VisitRadIfStmtRadStmt(ifStmt RadIfStmt) {
	for _, caseStmt := range ifStmt.Cases {
		val := caseStmt.Condition.Accept(r.i)
		if bval, ok := val.(bool); ok {
			if bval {
				for _, stmt := range caseStmt.Body {
					stmt.Accept(r)
				}
				return
			}
		} else {
			r.i.error(caseStmt.IfToken, fmt.Sprintf("If condition must be a boolean, got %s", TypeAsString(val)))
		}
	}

	if ifStmt.ElseBlock != nil {
		for _, stmt := range *ifStmt.ElseBlock {
			stmt.Accept(r)
		}
	}
}

// == radInvocationOld ==
