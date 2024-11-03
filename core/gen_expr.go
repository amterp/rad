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
	VisitMapExprExpr(MapExpr) interface{}
	VisitCollectionAccessExpr(CollectionAccess) interface{}
	VisitSliceAccessExpr(SliceAccess) interface{}
	VisitFunctionCallExpr(FunctionCall) interface{}
	VisitVariableExpr(Variable) interface{}
	VisitBinaryExpr(Binary) interface{}
	VisitTernaryExpr(Ternary) interface{}
	VisitLogicalExpr(Logical) interface{}
	VisitGroupingExpr(Grouping) interface{}
	VisitUnaryExpr(Unary) interface{}
	VisitListComprehensionExpr(ListComprehension) interface{}
	VisitVarPathExpr(VarPath) interface{}
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

type MapExpr struct {
	Keys           []Expr
	Values         []Expr
	OpenBraceToken Token
}

func (e MapExpr) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitMapExprExpr(e)
}
func (e MapExpr) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Keys: %v", e.Keys))
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	parts = append(parts, fmt.Sprintf("OpenBraceToken: %v", e.OpenBraceToken))
	return fmt.Sprintf("MapExpr(%s)", strings.Join(parts, ", "))
}

type CollectionAccess struct {
	Collection       Expr
	Key              Expr
	OpenBracketToken Token
}

func (e CollectionAccess) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitCollectionAccessExpr(e)
}
func (e CollectionAccess) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Collection: %v", e.Collection))
	parts = append(parts, fmt.Sprintf("Key: %v", e.Key))
	parts = append(parts, fmt.Sprintf("OpenBracketToken: %v", e.OpenBracketToken))
	return fmt.Sprintf("CollectionAccess(%s)", strings.Join(parts, ", "))
}

type SliceAccess struct {
	ListOrString     Expr
	OpenBracketToken Token
	Start            *Expr
	ColonToken       Token
	End              *Expr
}

func (e SliceAccess) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitSliceAccessExpr(e)
}
func (e SliceAccess) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("ListOrString: %v", e.ListOrString))
	parts = append(parts, fmt.Sprintf("OpenBracketToken: %v", e.OpenBracketToken))
	parts = append(parts, fmt.Sprintf("Start: %v", e.Start))
	parts = append(parts, fmt.Sprintf("ColonToken: %v", e.ColonToken))
	parts = append(parts, fmt.Sprintf("End: %v", e.End))
	return fmt.Sprintf("SliceAccess(%s)", strings.Join(parts, ", "))
}

type FunctionCall struct {
	Function                Token
	Args                    []Expr
	NamedArgs               []NamedArg
	NumExpectedReturnValues int
}

func (e FunctionCall) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitFunctionCallExpr(e)
}
func (e FunctionCall) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Function: %v", e.Function))
	parts = append(parts, fmt.Sprintf("Args: %v", e.Args))
	parts = append(parts, fmt.Sprintf("NamedArgs: %v", e.NamedArgs))
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

type Ternary struct {
	Condition    Expr
	QuestionMark Token
	True         Expr
	False        Expr
}

func (e Ternary) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitTernaryExpr(e)
}
func (e Ternary) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Condition: %v", e.Condition))
	parts = append(parts, fmt.Sprintf("QuestionMark: %v", e.QuestionMark))
	parts = append(parts, fmt.Sprintf("True: %v", e.True))
	parts = append(parts, fmt.Sprintf("False: %v", e.False))
	return fmt.Sprintf("Ternary(%s)", strings.Join(parts, ", "))
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

type ListComprehension struct {
	Expression  Expr
	For         Token
	Identifier1 Token
	Identifier2 *Token
	Range       Expr
	Condition   *Expr
}

func (e ListComprehension) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitListComprehensionExpr(e)
}
func (e ListComprehension) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Expression: %v", e.Expression))
	parts = append(parts, fmt.Sprintf("For: %v", e.For))
	parts = append(parts, fmt.Sprintf("Identifier1: %v", e.Identifier1))
	parts = append(parts, fmt.Sprintf("Identifier2: %v", e.Identifier2))
	parts = append(parts, fmt.Sprintf("Range: %v", e.Range))
	parts = append(parts, fmt.Sprintf("Condition: %v", e.Condition))
	return fmt.Sprintf("ListComprehension(%s)", strings.Join(parts, ", "))
}

type VarPath struct {
	Identifier Token
	Keys       []Expr
}

func (e VarPath) Accept(visitor ExprVisitor) interface{} {
	return visitor.VisitVarPathExpr(e)
}
func (e VarPath) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Identifier: %v", e.Identifier))
	parts = append(parts, fmt.Sprintf("Keys: %v", e.Keys))
	return fmt.Sprintf("VarPath(%s)", strings.Join(parts, ", "))
}
