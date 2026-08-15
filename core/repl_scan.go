package core

import (
	"strings"

	"github.com/amterp/rad/rts/radfmt"
)

// This is the REPL's whole understanding of Rad's lexical structure: enough to
// tell "you are still typing" from "that is a syntax error", and to guess the
// next line's indent. It is a hand-written scanner rather than a parse, for two
// reasons.
//
// The first is that a parse cannot answer the question. tree-sitter recovers
// from errors by relocating and shrinking the error region, so an incomplete
// `if true:` and outright nonsense like `@@@` both produce one ERROR node
// reaching the end of the input. There is no span test that separates them.
//
// The second is that we need the answer *affirmatively* anyway. Continuing on
// "something is unfinished" is the safe direction only when we can point at what
// is unfinished - an unclosed bracket, an unterminated string, an open block.
// Guessing "probably incomplete" from a failed parse strands the user in a
// prompt no keystroke will close.

// replScan is the lexical state of a buffer: what is still open, and what the
// code looks like with comments and string contents removed.
type replScan struct {
	depth      int      // unclosed ( [ {
	openString bool     // a string literal is still waiting for its delimiter
	codeLines  []string // one per line, strings and comments blanked out
}

// ReplNeedsMore reports whether a buffer is an unfinished Rad fragment that
// should keep accumulating.
//
// Callers must supply their own terminator - a blank line - because the block
// rule below never releases on its own. That is deliberate: an indentation
// scoped language has no closing token, so "the user typed nothing" is the only
// signal available, and both input paths implement it.
func ReplNeedsMore(buf string) bool {
	sc := scanRepl(buf)
	if sc.openString || sc.depth > 0 {
		return true
	}
	return sc.blockOpen()
}

// ReplAutoIndent returns the leading whitespace for the line after buf: one
// level deeper following a block header, otherwise level with the line above.
func ReplAutoIndent(buf string) string {
	sc := scanRepl(buf)
	last := sc.lastCodeLine()
	indent := last[:len(last)-len(strings.TrimLeft(last, " \t"))]
	if endsBlockHeader(last) {
		indent += radfmt.IndentUnit
	}
	return indent
}

// blockOpen reports whether any line has opened a block. It latches: once a
// block is open only a blank line closes it, because the body's own lines say
// nothing about whether the user is finished with them. Without the latch, a
// two-statement body would be impossible - the second statement's line does not
// end in a colon, so the buffer would submit after the first.
func (sc replScan) blockOpen() bool {
	for _, line := range sc.codeLines {
		if endsBlockHeader(line) {
			return true
		}
	}
	return false
}

func (sc replScan) lastCodeLine() string {
	if len(sc.codeLines) == 0 {
		return ""
	}
	return sc.codeLines[len(sc.codeLines)-1]
}

// endsBlockHeader reports whether a line of code opens a block. Every one of
// Rad's block forms - if, else, for, while, fn, switch, and the suffix catch -
// ends its header with a colon, and a comment may not follow it, so a trailing
// colon is the reliable signal. A colon elsewhere on the line (a map entry, a
// ternary, an inline lambda) does not end it.
func endsBlockHeader(line string) bool {
	return strings.HasSuffix(strings.TrimRight(line, " \t"), ":")
}

// ReplTrimTrailingBlankLines drops the whitespace-only lines a submitted buffer
// carries - the blank line that closed a block, and any auto-indent left on it.
// Rad's parser would otherwise see a dangling indent and reject the whole thing.
func ReplTrimTrailingBlankLines(buf string) string {
	lines := strings.Split(buf, "\n")
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	return strings.Join(lines, "\n")
}

// scanRepl walks the buffer once, tracking what is open and emitting each line
// with comments dropped and string literals reduced to a placeholder. Blanking
// literals is the point: a colon or bracket inside a string must not be read as
// structure, and Rad's interpolation puts real braces inside strings.
func scanRepl(src string) replScan {
	var sc replScan
	var line strings.Builder

	rs := []rune(src)
	var closer []rune // the delimiter the open string is waiting for
	raw := false

	flush := func() {
		sc.codeLines = append(sc.codeLines, line.String())
		line.Reset()
	}

	for i := 0; i < len(rs); {
		r := rs[i]

		if closer != nil {
			// A backslash escapes the next rune, so an escaped delimiter does not
			// close the string. Raw strings have no escapes at all.
			if !raw && r == '\\' && i+1 < len(rs) {
				i += 2
				continue
			}
			if matchesAt(rs, i, closer) {
				i += len(closer)
				closer = nil
				line.WriteRune('"') // the whole literal, as one inert rune
				continue
			}
			if r == '\n' {
				flush() // a multi-line string still spans lines
			}
			i++
			continue
		}

		switch {
		case r == '\n':
			flush()
			i++

		case r == '/' && i+1 < len(rs) && rs[i+1] == '/':
			for i < len(rs) && rs[i] != '\n' {
				i++
			}

		case r == 'r' && i+1 < len(rs) && isQuote(rs[i+1]):
			closer, raw = openString(rs, i+1), true
			i += 1 + len(closer)

		case isQuote(r):
			closer, raw = openString(rs, i), false
			i += len(closer)

		default:
			switch r {
			case '(', '[', '{':
				sc.depth++
			case ')', ']', '}':
				if sc.depth > 0 {
					sc.depth--
				}
			}
			line.WriteRune(r)
			i++
		}
	}
	flush()

	sc.openString = closer != nil
	return sc
}

// openString returns the delimiter that will close the string starting at i.
// Only double quotes have a triple-quoted multi-line form.
func openString(rs []rune, i int) []rune {
	if rs[i] == '"' && matchesAt(rs, i, []rune(`"""`)) {
		return []rune(`"""`)
	}
	return []rune{rs[i]}
}

func isQuote(r rune) bool {
	return r == '"' || r == '\'' || r == '`'
}

func matchesAt(rs []rune, i int, want []rune) bool {
	if i+len(want) > len(rs) {
		return false
	}
	for j, w := range want {
		if rs[i+j] != w {
			return false
		}
	}
	return true
}
