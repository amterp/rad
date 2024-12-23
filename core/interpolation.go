package core

import (
	"fmt"
	"strings"
)

// performStringInterpolation replaces {variables} in a string with their values
func (l *LiteralInterpreter) performStringInterpolation(stringLiteral StringLiteral) RslString {
	var result strings.Builder
	for i, stringToken := range stringLiteral.Value {
		result.WriteString(stringToken.Literal)
		if i < len(stringLiteral.InlineExprs) {
			inlineExpr := stringLiteral.InlineExprs[i]
			exprVal := inlineExpr.Expression.Accept(l.i)
			// todo bad token for sake of error
			formatted := l.format(stringToken, exprVal, inlineExpr.Formatting)
			result.WriteString(formatted)
		}
	}
	return NewRslString(result.String())
}

func (l *LiteralInterpreter) format(token Token, val interface{}, formatting *InlineExprFormat) string {
	if formatting == nil {
		return ToPrintable(val)
	}

	formatInfo := *formatting
	// expect GoFormat to not have any type specifier on the end yet from parser
	switch coerced := val.(type) {
	case int64:
		if formatInfo.IsFloatFormat {
			formatStr := l.insertWidthIntoNonStrFormat(formatInfo)
			formatStr += "f"
			return fmt.Sprintf(formatStr, float64(coerced))
		} else {
			formatStr := l.insertWidthIntoNonStrFormat(formatInfo)
			formatStr += "d"
			return fmt.Sprintf(formatStr, coerced)
		}
	case float64:
		formatStr := l.insertWidthIntoNonStrFormat(formatInfo)
		formatStr += "f"
		formatted := fmt.Sprintf(formatStr, coerced)

		// todo: removing trailing zeros. Need to consider left/right padding scenarios.
		//if !formatInfo.IsFloatFormat {
		//  // No explicit precision set -- then Go defaults to 6. Remove trailing zeros.
		//	for strings.HasSuffix(formatted, "0") {
		//		formatted = formatted[:len(formatted)-1]
		//	}
		//}

		return formatted
	default:
		if formatInfo.IsFloatFormat {
			l.i.error(token, fmt.Sprintf("Expected number for format, got %T", val))
			panic(UNREACHABLE)
		}
		valStr := ToPrintable(val)
		formatStr := l.insertWidthIntoStrFormat(formatInfo, val, valStr)
		formatStr += "s"
		return fmt.Sprintf(formatStr, valStr)
	}
}

func (l *LiteralInterpreter) insertWidthIntoStrFormat(formatInfo InlineExprFormat, val interface{}, valStr string) string {
	if formatInfo.Width == nil {
		return formatInfo.GoFormat
	}

	width := *formatInfo.Width

	// consider if we need to adjust padding due to color
	if str, ok := val.(RslString); ok {
		plainLen := str.Len()
		coloredLen := int64(len(valStr))
		diff := coloredLen - plainLen
		width += diff
	}

	return strings.ReplaceAll(formatInfo.GoFormat, PADDING_CHAR, fmt.Sprintf("%d", width))
}

func (l *LiteralInterpreter) insertWidthIntoNonStrFormat(formatInfo InlineExprFormat) string {
	if formatInfo.Width == nil {
		return formatInfo.GoFormat
	}
	return strings.ReplaceAll(formatInfo.GoFormat, PADDING_CHAR, fmt.Sprintf("%d", *formatInfo.Width))
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
