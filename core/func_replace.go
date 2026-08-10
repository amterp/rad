package core

import (
	"regexp"
	"strings"

	"github.com/amterp/rad/rts/rl"
)

var FuncReplace = BuiltInFunc{
	Name: FUNC_REPLACE,
	Execute: func(f FuncInvocation) RadValue {
		original := f.GetStr("_original").Plain()
		find := f.GetStr("_find").Plain()
		replace := f.GetStr("_replace").Plain()

		if !f.GetBool("regex") {
			return f.Return(strings.ReplaceAll(original, find, replace))
		}

		re, err := regexp.Compile(find)
		if err != nil {
			return f.ReturnErrf(rl.ErrInvalidRegex, "Error compiling regex pattern: %s", err)
		}

		return f.Return(replaceAllExpanding(re, original, replace))
	},
}

// replaceAllExpanding rewrites every match of re in original, expanding group
// references in the replacement template.
//
// The obvious spelling is ReplaceAllStringFunc plus FindStringSubmatch on each
// match, which is what this used to be. That re-runs the pattern against the
// matched text in isolation, and an assertion that held in context can fail out
// of it: `\Bb(c)` matches "bc" inside "abc", but "bc" on its own starts at a
// word boundary, so the second match finds nothing and the groups come back
// empty. Taking the group offsets from the original match is both correct and
// one pass cheaper.
func replaceAllExpanding(re *regexp.Regexp, original, template string) string {
	matches := re.FindAllStringSubmatchIndex(original, -1)
	if matches == nil {
		return original
	}

	var sb strings.Builder
	sb.Grow(len(original))

	submatches := make([]string, 0, re.NumSubexp()+1)
	end := 0
	for _, match := range matches {
		sb.WriteString(original[end:match[0]])

		submatches = submatches[:0]
		for g := 0; g < len(match)/2; g++ {
			start, stop := match[2*g], match[2*g+1]
			if start < 0 {
				// A group inside an alternation branch that didn't run.
				submatches = append(submatches, "")
			} else {
				submatches = append(submatches, original[start:stop])
			}
		}

		sb.WriteString(expandGroupRefs(template, submatches))
		end = match[1]
	}
	sb.WriteString(original[end:])

	return sb.String()
}

// expandGroupRefs substitutes `$0`, `$1`, `${1}` and `$$` in a replacement
// template against a match's submatches (index 0 being the whole match).
//
// We don't use Go's own `Regexp.Expand`, because it reads `$name` as the
// longest run of word characters, which would break the documented
// `replace("Name: abc", "a(b)c", "$1o$1", regex=true)` -> "Name: bob" case:
// Go sees group "1o", finds nothing, and drops it. Rad's rule is that a bare
// `$N` consumes digits only.
//
// Nor do we substitute group-by-group with strings.ReplaceAll, which is what
// this used to do. That approach had the `$1` pass corrupt `$10`, and let a
// submatch whose own text contained `$2` get re-substituted by a later pass.
// A single left-to-right scan is immune to both: every byte of output is
// written once and never revisited.
//
// A reference to a group that doesn't exist is left as written. Go's Expand
// would emit an empty string, but leaving it intact is the friendlier answer
// for the case this function exists to protect - replacement text relayed from
// somewhere else, where a stray `$` means nothing and should survive.
func expandGroupRefs(template string, submatches []string) string {
	var sb strings.Builder
	sb.Grow(len(template))

	for i := 0; i < len(template); {
		if template[i] != '$' {
			sb.WriteByte(template[i])
			i++
			continue
		}

		ref, width := parseGroupRef(template[i:])
		if width == 0 {
			sb.WriteByte('$')
			i++
			continue
		}
		if ref == groupRefEscapedDollar {
			sb.WriteByte('$')
		} else if ref < len(submatches) {
			sb.WriteString(submatches[ref])
		} else {
			sb.WriteString(template[i : i+width])
		}
		i += width
	}

	return sb.String()
}

const groupRefEscapedDollar = -1

// parseGroupRef reads a `$`-reference off the front of s, returning the group
// index it names and how many bytes it spans. A width of 0 means s doesn't
// start with a reference we recognize, and the `$` is literal text.
func parseGroupRef(s string) (group int, width int) {
	if len(s) < 2 {
		return 0, 0
	}

	if s[1] == '$' {
		return groupRefEscapedDollar, 2
	}

	braced := s[1] == '{'
	digitsStart := 1
	if braced {
		digitsStart = 2
	}

	digitsEnd := digitsStart
	for digitsEnd < len(s) && s[digitsEnd] >= '0' && s[digitsEnd] <= '9' {
		digitsEnd++
	}
	if digitsEnd == digitsStart {
		return 0, 0
	}

	width = digitsEnd
	if braced {
		if digitsEnd >= len(s) || s[digitsEnd] != '}' {
			return 0, 0
		}
		width = digitsEnd + 1
	}

	// A digit run this long can't name a real group, and parsing it risks
	// overflow. Treat the whole thing as literal text.
	if digitsEnd-digitsStart > 9 {
		return 0, 0
	}

	for _, c := range []byte(s[digitsStart:digitsEnd]) {
		group = group*10 + int(c-'0')
	}
	return group, width
}
