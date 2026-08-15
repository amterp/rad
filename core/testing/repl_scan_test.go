package testing

import (
	"testing"

	"github.com/amterp/rad/core"
	"github.com/stretchr/testify/assert"
)

// The REPL's continuation rule, which decides whether Enter runs your input or
// opens another line. These cases are the contract; the scanner is free to
// change shape underneath them.
func TestReplNeedsMore(t *testing.T) {
	cases := []struct {
		name string
		buf  string
		want bool
	}{
		// Complete input runs.
		{"assignment", `x = 1`, false},
		{"expression", `1 + 2`, false},
		{"call", `print("hi")`, false},
		{"balanced brackets", `f([1, 2], {"a": 1})`, false},
		{"map literal", `m = {"a": 1, "b": 2}`, false},
		{"empty", ``, false},

		// Unclosed brackets keep the buffer open.
		{"open paren", `f(1,`, true},
		{"open list", `x = [1, 2,`, true},
		{"open map", `m = {"a":`, true},
		{"nested still open", `f(g(1)`, true},

		// Unterminated strings keep it open.
		{"open double quote", `x = "abc`, true},
		{"open single quote", `x = 'abc`, true},
		{"open backtick", "x = `abc", true},
		{"open triple quote", `x = """`, true},
		{"closed triple quote", "x = \"\"\"\nbody\n\"\"\"", false},
		{"raw string open", `x = r"abc`, true},
		{"raw string closed", `x = r"abc"`, false},

		// A quote inside a string is not a delimiter.
		{"escaped quote", `x = "a\"b"`, false},
		{"raw string ignores escapes", `x = r"a\"`, false},

		// Brackets and colons inside strings are text, not structure. This is
		// why the predicate scans rather than pattern-matching.
		{"bracket in string", `x = "[["`, false},
		{"colon in string", `x = "if a:"`, false},
		{"interpolation braces balance", `x = "a{b}c"`, false},

		// Comments are not code.
		{"trailing comment", `x = 1 // note`, false},
		{"colon inside comment", `x = 1 // if a:`, false},
		{"comment only", `// just a note`, false},

		// Block headers open a block.
		{"if header", `if x > 1:`, true},
		{"for header", `for i in range(3):`, true},
		{"while header", `while true:`, true},
		{"fn header", `fn greet(name):`, true},
		{"switch header", `result = switch x:`, true},
		{"catch suffix", "$`ls` catch:", true},

		// The latch: a block stays open across its body, or a two-statement
		// body would be impossible to type.
		{"body after header", "if x:\n    print(1)", true},
		{"two statements in body", "if x:\n    print(1)\n    print(2)", true},
		{"nested blocks", "if a:\n    if b:\n        print(1)", true},
		{"else branch", "if a:\n    print(1)\nelse:\n    print(2)", true},

		// A colon that does not end a line is not a block header.
		{"ternary", `x = c ? "a" : "b"`, false},
		{"inline lambda", `f = fn(x): x + 1`, false},
		{"switch case arrow", "y = switch x:\n    case 1 -> 2", true},

		// Nonsense is a syntax error, not an invitation to keep typing. This is
		// the case a parse-based predicate gets wrong.
		{"garbage", `@@@`, false},
		{"stray close paren", `x = )`, false},
		{"dangling operator", `x = 1 +`, false},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, core.ReplNeedsMore(c.buf), "buffer: %q", c.buf)
		})
	}
}

func TestReplAutoIndent(t *testing.T) {
	cases := []struct {
		name string
		buf  string
		want string
	}{
		{"top level stays flush", `x = 1`, ""},
		{"block header indents", `if x:`, "    "},
		{"body keeps its indent", "if x:\n    print(1)", "    "},
		{"nested header indents again", "if a:\n    if b:", "        "},
		{"dedented else re-indents", "if a:\n    print(1)\nelse:", "    "},
		{"colon in a string does not indent", `x = "a:"`, ""},
		{"comment after code", `x = 1 // note`, ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, core.ReplAutoIndent(c.buf), "buffer: %q", c.buf)
		})
	}
}

// A submitted buffer carries the blank line that closed its block, plus any
// auto-indent left sitting on it. Rad's parser rejects a dangling indent, so
// the REPL trims before it parses.
func TestReplTrimTrailingBlankLines(t *testing.T) {
	cases := []struct {
		name string
		buf  string
		want string
	}{
		{"trailing indent-only line", "if x:\n    pass\n    ", "if x:\n    pass"},
		{"several blank lines", "x = 1\n\n\n", "x = 1"},
		{"nothing to trim", "x = 1", "x = 1"},
		{"interior blanks survive", "x = 1\n\ny = 2", "x = 1\n\ny = 2"},
		{"all blank", "\n  \n", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			assert.Equal(t, c.want, core.ReplTrimTrailingBlankLines(c.buf))
		})
	}
}
