// GENERATED -- DO NOT EDIT
package core

import (
	"fmt"
	"strings"
)

type ArrayLiteral interface {
	Accept(visitor ArrayLiteralVisitor) interface{}
}
type ArrayLiteralVisitor interface {
	VisitStringArrayLiteralArrayLiteral(StringArrayLiteral) interface{}
	VisitIntArrayLiteralArrayLiteral(IntArrayLiteral) interface{}
	VisitFloatArrayLiteralArrayLiteral(FloatArrayLiteral) interface{}
	VisitBoolArrayLiteralArrayLiteral(BoolArrayLiteral) interface{}
	VisitEmptyArrayLiteralArrayLiteral(EmptyArrayLiteral) interface{}
}
type StringArrayLiteral struct {
	Values []StringLiteral
}

func (e StringArrayLiteral) Accept(visitor ArrayLiteralVisitor) interface{} {
	return visitor.VisitStringArrayLiteralArrayLiteral(e)
}
func (e StringArrayLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("StringArrayLiteral(%s)", strings.Join(parts, ", "))
}

type IntArrayLiteral struct {
	Values []IntLiteral
}

func (e IntArrayLiteral) Accept(visitor ArrayLiteralVisitor) interface{} {
	return visitor.VisitIntArrayLiteralArrayLiteral(e)
}
func (e IntArrayLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("IntArrayLiteral(%s)", strings.Join(parts, ", "))
}

type FloatArrayLiteral struct {
	Values []FloatLiteral
}

func (e FloatArrayLiteral) Accept(visitor ArrayLiteralVisitor) interface{} {
	return visitor.VisitFloatArrayLiteralArrayLiteral(e)
}
func (e FloatArrayLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("FloatArrayLiteral(%s)", strings.Join(parts, ", "))
}

type BoolArrayLiteral struct {
	Values []BoolLiteral
}

func (e BoolArrayLiteral) Accept(visitor ArrayLiteralVisitor) interface{} {
	return visitor.VisitBoolArrayLiteralArrayLiteral(e)
}
func (e BoolArrayLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("BoolArrayLiteral(%s)", strings.Join(parts, ", "))
}

type EmptyArrayLiteral struct {
}

func (e EmptyArrayLiteral) Accept(visitor ArrayLiteralVisitor) interface{} {
	return visitor.VisitEmptyArrayLiteralArrayLiteral(e)
}
func (e EmptyArrayLiteral) String() string {
	return "EmptyArrayLiteral()"
}
