// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type LiteralOrArray interface {
	Accept(visitor LiteralOrArrayVisitor) interface{}
}
type LiteralOrArrayVisitor interface {
	VisitLoaLiteralLiteralOrArray(LoaLiteral) interface{}
	VisitLoaArrayLiteralOrArray(LoaArray) interface{}
}
type LoaLiteral struct {
	Value Literal
}

func (e LoaLiteral) Accept(visitor LiteralOrArrayVisitor) interface{} {
	return visitor.VisitLoaLiteralLiteralOrArray(e)
}
func (e LoaLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("LoaLiteral(%s)", strings.Join(parts, ", "))
}

type LoaArray struct {
	Value ArrayLiteral
}

func (e LoaArray) Accept(visitor LiteralOrArrayVisitor) interface{} {
	return visitor.VisitLoaArrayLiteralOrArray(e)
}
func (e LoaArray) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Value: %v", e.Value))
	return fmt.Sprintf("LoaArray(%s)", strings.Join(parts, ", "))
}
