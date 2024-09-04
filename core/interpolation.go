package core

import (
	"fmt"
	"strings"
)

// Helper function to process the string
func processString(
	s string,
	callback func(capturing bool, escaped bool, char rune, variable string, result *strings.Builder, env *Env), env *Env,
) string {
	var result strings.Builder
	capturing := false
	escaped := false
	variable := ""

	for _, char := range s {
		if char == '\\' {
			if escaped {
				result.WriteRune(char)
				escaped = false
			} else {
				escaped = true
			}
			continue
		}

		if char == '{' {
			if escaped {
				result.WriteRune(char)
				escaped = false
			} else if !capturing {
				capturing = true
				variable = ""
			} else {
				variable += string(char)
			}
			continue
		}

		if char == '}' {
			if escaped {
				result.WriteRune(char)
				escaped = false
			} else if capturing {
				capturing = false
				if variable != "" {
					callback(capturing, escaped, char, variable, &result, env)
				}
			} else {
				result.WriteRune(char)
			}
			continue
		}

		if capturing {
			variable += string(char)
		} else {
			result.WriteRune(char)
		}
	}

	return result.String()
}

// performStringInterpolation replaces {variables} in a string with their values
func performStringInterpolation(s string, env *Env) string {
	return processString(s, func(capturing bool, escaped bool, char rune, variable string, result *strings.Builder, env *Env) {
		value := env.GetByName(variable).value
		result.WriteString(fmt.Sprintf("%v", value))
	}, env)
}

// extractVariables extracts variables within non-escaped {} brackets in the input string
func extractVariables(s string) []string {
	var variables []string
	processString(s, func(capturing bool, escaped bool, char rune, variable string, result *strings.Builder, env *Env) {
		variables = append(variables, variable)
	}, nil)
	return variables
}
