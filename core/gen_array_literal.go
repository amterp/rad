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
	VisitMixedArrayLiteralArrayLiteral(MixedArrayLiteral) interface{}
}
type MixedArrayLiteral struct {
	Values []LiteralOrArray
}

func (e MixedArrayLiteral) Accept(visitor ArrayLiteralVisitor) interface{} {
	return visitor.VisitMixedArrayLiteralArrayLiteral(e)
}
func (e MixedArrayLiteral) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("Values: %v", e.Values))
	return fmt.Sprintf("MixedArrayLiteral(%s)", strings.Join(parts, ", "))
}
