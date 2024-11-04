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
	VisitColorRadFieldModStmt(Color)
	VisitMapModRadFieldModStmt(MapMod)
}
type Color struct {
	ColorToken Token
	ColorValue Expr
	Regex      Expr
}

func (e Color) Accept(visitor RadFieldModStmtVisitor) {
	visitor.VisitColorRadFieldModStmt(e)
}
func (e Color) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("ColorToken: %v", e.ColorToken))
	parts = append(parts, fmt.Sprintf("ColorValue: %v", e.ColorValue))
	parts = append(parts, fmt.Sprintf("Regex: %v", e.Regex))
	return fmt.Sprintf("Color(%s)", strings.Join(parts, ", "))
}

type MapMod struct {
	MapToken Token
	Op       Lambda
}

func (e MapMod) Accept(visitor RadFieldModStmtVisitor) {
	visitor.VisitMapModRadFieldModStmt(e)
}
func (e MapMod) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("MapToken: %v", e.MapToken))
	parts = append(parts, fmt.Sprintf("Op: %v", e.Op))
	return fmt.Sprintf("MapMod(%s)", strings.Join(parts, ", "))
}
