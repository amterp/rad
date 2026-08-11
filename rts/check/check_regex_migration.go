package check

import (
	"strings"

	"github.com/amterp/rad/rts/rl"
)

// Check 12: split()/replace() patterns that read as regexes but don't pass regex=true.
//
// This whole file is a migration aid. v0.12 flipped split() and replace() from
// regex-by-default to literal-by-default. A script written against the old
// default still parses and still runs - it just quietly stops matching and
// hands back the input unchanged. That silence is the entire cost of the flip,
// so this warning is on by default rather than living in strictOnlyCodes.
//
// Delete the file once v0.12 is a couple of releases behind us.

// regexMigrationFuncs maps a flipped builtin to the positional index of its
// pattern parameter, counting from the plain call form. Both are the second
// argument: split(_val, _sep) and replace(_original, _find, _replace).
var regexMigrationFuncs = map[string]int{
	"split":   1,
	"replace": 1,
}

func (c *RadCheckerImpl) addRegexPatternWithoutRegexArgWarnings(resolved *Resolved, d *[]Diagnostic) {
	if c.ast == nil || resolved == nil {
		return
	}

	ufcsCalls := collectUfcsCalls(c.ast)

	walkAST(c.ast, func(node rl.Node) {
		call, ok := node.(*rl.Call)
		if !ok {
			return
		}

		name, ok := builtinCallName(call, resolved)
		if !ok {
			return
		}
		patternIdx, ok := regexMigrationFuncs[name]
		if !ok {
			return
		}
		if ufcsCalls[call] {
			// `s.split(sep)` - the receiver fills the first parameter, so
			// every later one shifts down by an index.
			patternIdx--
		}
		if patternIdx < 0 || patternIdx >= len(call.Args) {
			return
		}
		if callHasNamedArg(call, "regex") {
			return
		}

		pattern, ok := call.Args[patternIdx].(*rl.LitString)
		if !ok {
			return
		}
		text := patternText(pattern)
		signal, found := regexSignal(text)
		if !found {
			return
		}

		msg := "'" + truncate(text, 20) + "' reads as a regex ('" + signal + "'), but " +
			name + "() matches it literally"
		// "Silently" is the load-bearing word and stays inline: the failure mode
		// is a wrong answer with no error, so a reader who skims this and moves
		// on has no second chance to notice.
		suggestion := "Pass regex=true, or ignore this if the text really is literal - " +
			"otherwise it silently matches nothing. See https://amterp.dev/rad/migrations/v0.12/"

		diag := NewDiagnosticWarnFromSpan(pattern.Span(), c.src, msg, rl.ErrRegexPatternWithoutRegexArg)
		diag.Suggestion = &suggestion
		*d = append(*d, diag)
	})
}

// patternText returns the statically-known text of a string literal. For an
// interpolated string it returns the literal segments run together, dropping the
// interpolations. That's not the runtime value, but a pattern assembled from
// pieces is if anything *more* likely to be a regex, and its fixed parts carry
// the tell-tale constructs just the same.
func patternText(lit *rl.LitString) string {
	if lit.Simple {
		return lit.Value
	}
	var sb strings.Builder
	for _, seg := range lit.Segments {
		if seg.IsLiteral {
			sb.WriteString(seg.Text)
		}
	}
	return sb.String()
}

// collectUfcsCalls finds every Call reached through UFCS syntax (`s.split(x)`),
// which the converter nests inside a VarPath segment rather than leaving at the
// top level. Their argument lists are one short, because the receiver is
// implicit.
func collectUfcsCalls(ast rl.Node) map[*rl.Call]bool {
	ufcs := make(map[*rl.Call]bool)
	walkAST(ast, func(node rl.Node) {
		path, ok := node.(*rl.VarPath)
		if !ok {
			return
		}
		for _, seg := range path.Segments {
			if !seg.IsUFCS {
				continue
			}
			if call, ok := seg.Index.(*rl.Call); ok {
				ufcs[call] = true
			}
		}
	})
	return ufcs
}

// builtinCallName returns the name of the builtin a call targets. A user-defined
// function shadowing a builtin name resolves to its own symbol, so it won't match.
func builtinCallName(call *rl.Call, resolved *Resolved) (string, bool) {
	ident, ok := call.Func.(*rl.Identifier)
	if !ok {
		return "", false
	}
	sym, ok := resolved.Uses[ident]
	if !ok || sym == nil || sym.Kind != SymBuiltin {
		return "", false
	}
	return sym.Name, true
}

func callHasNamedArg(call *rl.Call, name string) bool {
	for _, arg := range call.NamedArgs {
		if arg.Name == name {
			return true
		}
	}
	return false
}

// regexEscapeSignals are the characters that, when backslash-escaped, mean the
// author was writing a regex. The classes and anchors (\d, \b, \A) have no
// literal reading at all; the metacharacters (\., \+) would be a literal
// backslash followed by that character, which nobody writes on purpose.
const regexEscapeSignals = `dswbDSWBAzZ.+*?()[]{}|^$`

// regexSignal reports whether a pattern contains a construct that only makes
// sense if it was written as a regex, and returns that construct for the
// message.
//
// Deliberately conservative, because a warning that cries wolf gets ignored and
// then the real migration hits go unnoticed. A bare '.' is the most common
// metacharacter in ordinary text - file extensions, version numbers, prose -
// and under the new literal default it needs no escaping, so flagging it would
// bury every real hit. Same reasoning for a lone '|' or '+', which are ordinary
// separators. Every rule below demands structure rather than a bare
// metacharacter sighting.
func regexSignal(pattern string) (string, bool) {
	if len(pattern) < 2 {
		return "", false
	}

	openBracket, openParen := -1, -1
	for i := 0; i < len(pattern); i++ {
		ch := pattern[i]

		if ch == '\\' && i+1 < len(pattern) {
			next := pattern[i+1]
			if strings.IndexByte(regexEscapeSignals, next) >= 0 {
				return `\` + string(next), true
			}
			// Some other escape: consume both bytes so the escaped
			// character isn't judged on its own.
			i++
			continue
		}

		switch ch {
		case '[':
			openBracket = i
		case ']':
			if openBracket >= 0 && i > openBracket+1 {
				return "[...]", true
			}
		case '(':
			openParen = i
		case ')':
			if openParen >= 0 {
				return "(...)", true
			}
		case '+', '*', '?':
			// A quantifier needs something to quantify, so one at the very
			// start is just that character.
			if i > 0 {
				return string(ch), true
			}
		case '|':
			// Alternation needs branches on both sides. A leading or trailing
			// '|' is far more likely to be a literal pipe.
			if i > 0 && i < len(pattern)-1 {
				return "|", true
			}
		case '{':
			if i > 0 {
				if width := repetitionWidth(pattern[i:]); width > 0 {
					return pattern[i : i+width], true
				}
			}
		}
	}

	// Anchors last: they're the weakest signal, and reporting '$' for a
	// pattern that also contains '\s' would name the least interesting half of
	// why it fired. An escaped trailing '$' is a literal dollar, and the scan
	// above has already judged it as an escape.
	if pattern[0] == '^' {
		return "^", true
	}
	if pattern[len(pattern)-1] == '$' && pattern[len(pattern)-2] != '\\' {
		return "$", true
	}

	return "", false
}

// repetitionWidth returns the byte length of a `{n}`, `{n,}` or `{n,m}`
// repetition at the front of s, or 0 if there isn't one.
func repetitionWidth(s string) int {
	i := 1 // past '{'
	digits := 0
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
		digits++
	}
	if digits == 0 {
		return 0
	}
	if i < len(s) && s[i] == ',' {
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	if i < len(s) && s[i] == '}' {
		return i + 1
	}
	return 0
}
