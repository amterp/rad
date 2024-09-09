package core

import (
	"fmt"
	"regexp"
	"strings"
)

// Replace allows capture group replacing, for example
// replace("Name: abc", "a(b)c", "$1o$1") will return "Name: bobby"
func Replace(i *MainInterpreter, function Token, oldString string, regexForOld string, regexForNew string) string {
	re, err := regexp.Compile(regexForOld)
	if err != nil {
		i.error(function, fmt.Sprintf("Error compiling regex pattern: %s", err))
	}

	replacementFunc := func(match string) string {
		submatches := re.FindStringSubmatch(match)

		if len(submatches) == 0 {
			return match
		}

		result := regexForNew
		for i, submatch := range submatches {
			placeholder := fmt.Sprintf("$%d", i)
			result = strings.ReplaceAll(result, placeholder, submatch)
		}

		return result
	}

	newString := re.ReplaceAllStringFunc(oldString, replacementFunc)

	return newString
}
