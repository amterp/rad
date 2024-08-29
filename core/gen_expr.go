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
	VisitLiteralExprExpr(LiteralExpr) interface{}
	VisitExprLoaExpr(ExprLoa) interface{}
	VisitArrayExprExpr(ArrayExpr) interface{}
	VisitArrayAccessExpr(ArrayAccess) interface{}
	VisitFunctionCallExpr(FunctionCall) interface{}
	VisitVariableExpr(Variable) interface{}
}
type LiteralExpr struct {
	Value Literal
}

func (e LiteralExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitLiteralExprExpr(e)
}
func (e LiteralExpr) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("LiteralExpr(%s)", strings.Join(parts, ", "))
}

type ExprLoa struct {
	Value LiteralOrArray
}

func (e ExprLoa) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitExprLoaExpr(e)
}
func (e ExprLoa) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("ExprLoa(%s)", strings.Join(parts, ", "))
}

type ArrayExpr struct {
	Values []Expr
}

func (e ArrayExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitArrayExprExpr(e)
}
func (e ArrayExpr) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("ArrayExpr(%s)", strings.Join(parts, ", "))
}

type ArrayAccess struct {
	Array Expr
	Index Expr
}

func (e ArrayAccess) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitArrayAccessExpr(e)
}
func (e ArrayAccess) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Array: %v", e.Array))
	parts = append(parts, fmt.Sprintf("Index: %v", e.Index))
	return fmt.Sprintf("ArrayAccess(%s)", strings.Join(parts, ", "))
}

type FunctionCall struct {
	Function Token
}

func (e FunctionCall) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitFunctionCallExpr(e)
}
func (e FunctionCall) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Function: %v", e.Function))
	return fmt.Sprintf("FunctionCall(%s)", strings.Join(parts, ", "))
}

type Variable struct {
	Name Token
}

func (e Variable) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitVariableExpr(e)
}
func (e Variable) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Name: %v", e.Name))
	return fmt.Sprintf("Variable(%s)", strings.Join(parts, ", "))
}
