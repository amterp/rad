package core

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/amterp/rad/rts/rl"
)

// Allows capture group replacing, for example
// replace("Name: abc", "a(b)c", "$1o$1") will return "Name: bobby"
var FuncReplace = BuiltInFunc{
	Name: FUNC_REPLACE,
	Execute: func(f FuncInvocation) RadValue {
		original := f.GetStr("_original").Plain()
		find := f.GetStr("_find").Plain()
		replace := f.GetStr("_replace").Plain()

		re, err := regexp.Compile(find)
		if err != nil {
			return f.ReturnErrf(rl.ErrInvalidRegex, "Error compiling regex pattern: %s", err)
		}

		replacementFunc := func(match string) string {
			submatches := re.FindStringSubmatch(match)

			if len(submatches) == 0 {
				return match
			}

			result := replace
			for i, submatch := range submatches {
				placeholder := fmt.Sprintf("$%d", i)
				result = strings.ReplaceAll(result, placeholder, submatch)
			}

			return result
		}

		newString := re.ReplaceAllStringFunc(original, replacementFunc)

		return f.Return(newString)
	},
}
