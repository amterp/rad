// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type Literal interface {
	Accept(visitor LiteralVisitor) interface{}
}
type LiteralVisitor interface {
	VisitStringLiteralLiteral(StringLiteral) interface{}
	VisitIntLiteralLiteral(IntLiteral) interface{}
	VisitFloatLiteralLiteral(FloatLiteral) interface{}
	VisitBoolLiteralLiteral(BoolLiteral) interface{}
}
type StringLiteral struct {
	Value       []StringLiteralToken
	InlineExprs []InlineExpr
}

func (e StringLiteral) Accept(visitor LiteralVisitor) interface{} {
	return visitor.VisitStringLiteralLiteral(e)
}
func (e StringLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	parts = append(parts, fmt.Sprintf("InlineExprs: %v", e.InlineExprs))
	return fmt.Sprintf("StringLiteral(%s)", strings.Join(parts, ", "))
}

type IntLiteral struct {
	Value IntLiteralToken
}

func (e IntLiteral) Accept(visitor LiteralVisitor) interface{} {
	return visitor.VisitIntLiteralLiteral(e)
}
func (e IntLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("IntLiteral(%s)", strings.Join(parts, ", "))
}

type FloatLiteral struct {
	Value FloatLiteralToken
}

func (e FloatLiteral) Accept(visitor LiteralVisitor) interface{} {
	return visitor.VisitFloatLiteralLiteral(e)
}
func (e FloatLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("FloatLiteral(%s)", strings.Join(parts, ", "))
}

type BoolLiteral struct {
	Value BoolLiteralToken
}

func (e BoolLiteral) Accept(visitor LiteralVisitor) interface{} {
	return visitor.VisitBoolLiteralLiteral(e)
}
func (e BoolLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("BoolLiteral(%s)", strings.Join(parts, ", "))
}
