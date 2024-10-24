package core

import (
	"fmt"
	"strings"
)

// performStringInterpolation replaces {variables} in a string with their values
func (l *LiteralInterpreter) performStringInterpolation(stringLiteral StringLiteral) string {
	var result strings.Builder
	for i, stringToken := range stringLiteral.Value {
		result.WriteString(stringToken.Literal)
		if i < len(stringLiteral.InlineExprs) {
			inlineExpr := stringLiteral.InlineExprs[i]
			exprVal := inlineExpr.Expression.Accept(l.i)
			result.WriteString(ToPrintable(exprVal))
			// todo format
		}
	}
	return result.String()
}

// extractVariables extracts variables within non-escaped {} brackets in the input string
// big oof on this implementation, think about how to improve
func extractVariables(i *MainInterpreter, blockToken Token, expr Expr) []string {
	var varNames []string
	switch coerced := expr.(type) {
	case ExprLoa:
		switch coerced2 := coerced.Value.(type) {
		case *LoaLiteral:
			switch coerced3 := coerced2.Value.(type) {
			case StringLiteral:
				inlineExprs := coerced3.InlineExprs
				for _, inlineExpr := range inlineExprs {
					expr := inlineExpr.Expression
					if variable, ok := expr.(*Variable); ok {
						varNames = append(varNames, variable.Name.GetLexeme())
					}
				}
			default:
				i.error(blockToken, fmt.Sprintf("Bug! Expected only string literals in block cases, got %T", coerced3))
			}
		default:
			i.error(blockToken, fmt.Sprintf("Bug! Expected only string literals in block cases, got %T", coerced2))
		}
	default:
		i.error(blockToken, fmt.Sprintf("Bug! Expected only string literals in block cases, got %T", coerced))
	}
	return varNames
}
