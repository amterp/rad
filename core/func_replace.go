package core

import (
	"fmt"
	"regexp"
	"strings"
)

// Allows capture group replacing, for example
// replace("Name: abc", "a(b)c", "$1o$1") will return "Name: bobby"
var FuncReplace = BuiltInFunc{
	Name: FUNC_REPLACE,
	Execute: func(f FuncInvocation) RadValue {
		oldStringArg := f.args[0]
		regexForOldArg := f.args[1]
		regexForNewArg := f.args[2]

		oldString := oldStringArg.value.RequireStr(f.i, oldStringArg.node).Plain()
		regexForOld := regexForOldArg.value.RequireStr(f.i, regexForOldArg.node).Plain()
		regexForNew := regexForNewArg.value.RequireStr(f.i, regexForNewArg.node).Plain()

		re, err := regexp.Compile(regexForOld)
		if err != nil {
			f.i.errorf(regexForOldArg.node, fmt.Sprintf("Error compiling regex pattern: %s", err))
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

		return newRadValues(f.i, f.callNode, newString)
	},
}
