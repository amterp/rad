// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type Expr interface {
	Accept(visitor ExprVisitor) interface{}
}
type ExprVisitor interface {
	VisitStringLiteralExpr(*StringLiteral) interface{}
	VisitIntLiteralExpr(*IntLiteral) interface{}
	VisitFloatLiteralExpr(*FloatLiteral) interface{}
	VisitBoolLiteralExpr(*BoolLiteral) interface{}
	VisitVariableExpr(*Variable) interface{}
}
type StringLiteral struct {
	value Token
}

func (e *StringLiteral) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitStringLiteralExpr(e)
}
func (e *StringLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("value: %v", e.value))
	return fmt.Sprintf("StringLiteral(%s)", strings.Join(parts, ", "))
}

type IntLiteral struct {
	value Token
}

func (e *IntLiteral) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitIntLiteralExpr(e)
}
func (e *IntLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("value: %v", e.value))
	return fmt.Sprintf("IntLiteral(%s)", strings.Join(parts, ", "))
}

type FloatLiteral struct {
	value Token
}

func (e *FloatLiteral) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitFloatLiteralExpr(e)
}
func (e *FloatLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("value: %v", e.value))
	return fmt.Sprintf("FloatLiteral(%s)", strings.Join(parts, ", "))
}

type BoolLiteral struct {
	value Token
}

func (e *BoolLiteral) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitBoolLiteralExpr(e)
}
func (e *BoolLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("value: %v", e.value))
	return fmt.Sprintf("BoolLiteral(%s)", strings.Join(parts, ", "))
}

type Variable struct {
	name Token
}

func (e *Variable) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitVariableExpr(e)
}
func (e *Variable) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("name: %v", e.name))
	return fmt.Sprintf("Variable(%s)", strings.Join(parts, ", "))
}
