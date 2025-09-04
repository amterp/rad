package core

import (
	"regexp"

	"github.com/amterp/rad/rts/rl"
)

var FuncMatches = BuiltInFunc{
	Name: FUNC_MATCHES,
	Execute: func(f FuncInvocation) RadValue {
		input := f.GetStr("_str").Plain()
		pattern := f.GetStr("_pattern").Plain()
		partial := f.GetBool("partial")

		re, err := regexp.Compile(pattern)
		if err != nil {
			return f.ReturnErrf(rl.ErrInvalidRegex, "Error compiling regex pattern: %s", err)
		}

		var matches bool
		if partial {
			matches = re.FindString(input) != ""
		} else {
			// anchoring pattern to ensure patterns like cat|dog get handled correctly
			anchoredPattern := "^(?:" + pattern + ")$"
			anchoredRe, err := regexp.Compile(anchoredPattern)
			if err != nil {
				return f.ReturnErrf(rl.ErrInvalidRegex, "Error compiling regex pattern: %s", err)
			}
			matches = anchoredRe.MatchString(input)
		}

		return f.Return(matches)
	},
}
