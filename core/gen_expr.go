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
	VisitExprLoaExpr(ExprLoa) interface{}
	VisitArrayExprExpr(ArrayExpr) interface{}
	VisitArrayAccessExpr(ArrayAccess) interface{}
	VisitFunctionCallExpr(FunctionCall) interface{}
	VisitVariableExpr(Variable) interface{}
	VisitBinaryExpr(Binary) interface{}
	VisitLogicalExpr(Logical) interface{}
	VisitGroupingExpr(Grouping) interface{}
	VisitUnaryExpr(Unary) interface{}
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
	Array            Expr
	Index            Expr
	OpenBracketToken Token
}

func (e ArrayAccess) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitArrayAccessExpr(e)
}
func (e ArrayAccess) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Array: %v", e.Array))
	parts = append(parts, fmt.Sprintf("Index: %v", e.Index))
	parts = append(parts, fmt.Sprintf("OpenBracketToken: %v", e.OpenBracketToken))
	return fmt.Sprintf("ArrayAccess(%s)", strings.Join(parts, ", "))
}

type FunctionCall struct {
	Function                Token
	Args                    []Expr
	NumExpectedReturnValues int
}

func (e FunctionCall) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitFunctionCallExpr(e)
}
func (e FunctionCall) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Function: %v", e.Function))
	parts = append(parts, fmt.Sprintf("Args: %v", e.Args))
	parts = append(parts, fmt.Sprintf("NumExpectedReturnValues: %v", e.NumExpectedReturnValues))
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

type Binary struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (e Binary) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitBinaryExpr(e)
}
func (e Binary) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Left: %v", e.Left))
	parts = append(parts, fmt.Sprintf("Operator: %v", e.Operator))
	parts = append(parts, fmt.Sprintf("Right: %v", e.Right))
	return fmt.Sprintf("Binary(%s)", strings.Join(parts, ", "))
}

type Logical struct {
	Left     Expr
	Operator Token
	Right    Expr
}

func (e Logical) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitLogicalExpr(e)
}
func (e Logical) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Left: %v", e.Left))
	parts = append(parts, fmt.Sprintf("Operator: %v", e.Operator))
	parts = append(parts, fmt.Sprintf("Right: %v", e.Right))
	return fmt.Sprintf("Logical(%s)", strings.Join(parts, ", "))
}

type Grouping struct {
	Value Expr
}

func (e Grouping) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitGroupingExpr(e)
}
func (e Grouping) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("Grouping(%s)", strings.Join(parts, ", "))
}

type Unary struct {
	Operator Token
	Right    Expr
}

func (e Unary) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitUnaryExpr(e)
}
func (e Unary) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Operator: %v", e.Operator))
	parts = append(parts, fmt.Sprintf("Right: %v", e.Right))
	return fmt.Sprintf("Unary(%s)", strings.Join(parts, ", "))
}
