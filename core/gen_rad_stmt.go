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
	VisitSortRadStmt(Sort)
	VisitFieldModsRadStmt(FieldMods)
	VisitRadIfStmtRadStmt(RadIfStmt)
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

type Sort struct {
	SortToken   Token
	Identifiers []Token
	Directions  []SortDir
	GeneralSort *SortDir
}

func (e Sort) Accept(visitor RadStmtVisitor) {
	visitor.VisitSortRadStmt(e)
}
func (e Sort) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("SortToken: %v", e.SortToken))
	parts = append(parts, fmt.Sprintf("Identifiers: %v", e.Identifiers))
	parts = append(parts, fmt.Sprintf("Directions: %v", e.Directions))
	parts = append(parts, fmt.Sprintf("GeneralSort: %v", e.GeneralSort))
	return fmt.Sprintf("Sort(%s)", strings.Join(parts, ", "))
}

type FieldMods struct {
	Identifiers []Token
	Mods        []RadFieldModStmt
}

func (e FieldMods) Accept(visitor RadStmtVisitor) {
	visitor.VisitFieldModsRadStmt(e)
}
func (e FieldMods) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifiers: %v", e.Identifiers))
	parts = append(parts, fmt.Sprintf("Mods: %v", e.Mods))
	return fmt.Sprintf("FieldMods(%s)", strings.Join(parts, ", "))
}

type RadIfStmt struct {
	Cases     []RadIfCase
	ElseBlock *[]RadStmt
}

func (e RadIfStmt) Accept(visitor RadStmtVisitor) {
	visitor.VisitRadIfStmtRadStmt(e)
}
func (e RadIfStmt) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Cases: %v", e.Cases))
	parts = append(parts, fmt.Sprintf("ElseBlock: %v", e.ElseBlock))
	return fmt.Sprintf("RadIfStmt(%s)", strings.Join(parts, ", "))
}
