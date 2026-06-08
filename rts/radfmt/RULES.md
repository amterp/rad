# Rad Formatter Rules

This is the authoritative catalog of every formatting decision `rad fmt` makes.
It is the source of truth: the code enforces these rules, the snapshots
demonstrate them, and a coverage test (`rules_test.go`) keeps all three in sync.

## How to read this

Each rule has a stable ID (`F12`), a one-line statement, a `before → after`
example, and a pointer to the code that enforces it. Rule headings are
machine-parsed, so keep the shape exactly (real headings use a number; this
illustrative one uses `Fn` so the coverage test doesn't read it as a rule):

```
### Fn - Short rule title `status`
```

- **IDs are stable and append-only.** A new rule takes the next free number. A
  removed rule becomes a tombstone (see Changelog) - its number is never reused
  and never renumbered, so references in code, commits, and discussion stay
  valid forever.
- **Status** (the backticked word in each heading) drives enforcement:
  - `implemented` - an active canonicalizing rule. Must have a `// [Fn]` code
    tag and at least one snapshot whose title contains `[Fn]`.
  - `passthrough` - a deliberate non-action (we intentionally leave something
    alone). Same enforcement as `implemented` - tag + snapshot - because "we
    don't touch this" is a decision worth locking with a test.
  - `limitation` - a known gap we mean to close. Documented and code-tagged by
    convention (the tag is not enforced); exempt from the snapshot requirement.
  - `deferred` / `roadmap` - decided or planned but not built. Documented only.

Byte-level rules (line endings, trailing whitespace, the exact trailing newline)
can't be written as literal snapshot text - the test scanner strips CRs and
editors strip trailing spaces. Those cases mark their title with `[raw]` and
carry their input and expected output as Go-quoted strings (e.g. `"x = 1\r\n"`),
which the harness decodes and compares byte-for-byte. So every rule, including
these, is a real snapshot - there is no "trust me, it's tested elsewhere" status.

The target style derives from existing idiom (`core/embedded/*`) and the
decisions recorded in the git history of this package.

---

## Whitespace & file

### F1 - Four-space indentation `implemented`
Each block level indents four spaces; never tabs.
Code: `render.go` `IndentUnit`, `print_stmt.go` `indentedBody`

### F2 - Line endings normalized to LF `implemented`
CRLF and bare CR become `\n`.
`x = 1\r\n` → `x = 1\n`
Code: `fmt.go` `normalizeLineEndings`

### F3 - Exactly one trailing newline `implemented`
The file ends in a single `\n`, no more, no less.
Code: `printer.go` `formatSourceFile`

### F4 - Trailing whitespace stripped `implemented`
Every line has its trailing spaces/tabs removed.
Code: `render.go` `trimTrailing`

### F5 - Target line width 120 `implemented`
Calls and collections wrap when a line would exceed 120 columns. It is a
target, not a hard cap - long unbreakable tokens (strings, comments) may exceed
it.
Code: `render.go` `MaxWidth`

---

## Blank lines & comments

### F6 - Collapse multiple blank lines `implemented`
Two or more consecutive blank lines become one.
`a\n\n\n\nb` → `a\n\nb`
Code: `printer.go` `formatSeq`

### F7 - Strip blank lines at edges `implemented`
No blank lines at the start or end of a file or block body.
Code: `printer.go` `formatSeq`

### F8 - Preserve a single blank line `implemented`
One blank line is kept wherever the source had at least one, as a separator.
Code: `printer.go` `formatSeq`

### F9 - Standalone comment keeps its line `implemented`
A comment on its own line stays on its own line.
Code: `printer.go` `formatSeq`

### F10 - Trailing same-line comment stays `implemented`
A comment after code on the same line stays there (it attaches as a
line-suffix, so it never falls inside wrapped code).
`x = 1 // note` → `x = 1 // note`
Code: `printer.go` `formatSeq`

### F11 - Header-trailing comment stays on the header `implemented`
A comment trailing a block header (`if x: // why`) stays on the header line
rather than being pushed into the body.
Code: `print_stmt.go` `blockTail`

### F36 - Comment-bearing expressions emitted verbatim `limitation`
An expression containing an interior comment is emitted verbatim rather than
risk dropping the comment during reflow. To be removed as per-construct interior
comment attachment matures (see `DESIGN.md`).
Code: `print_expr.go` `formatExpr`

---

## Statements

### F12 - Assignment spacing `implemented`
One space on each side of `=`.
`a=1` → `a = 1`
Code: `print_stmt.go` `formatAssign`

### F13 - Multi-assignment spacing `implemented`
`", "` between targets; spaces around `=`.
`a,b=f()` → `a, b = f()`
Code: `print_stmt.go` `formatAssign`

### F14 - Compound assignment spacing `implemented`
One space around `+=`, `-=`, `*=`, `/=`, `%=`.
`x+=1` → `x += 1`
Code: `print_stmt.go` `formatCompoundAssign`

### F15 - Increment / decrement bind tight `implemented`
No inner space.
`i ++` → `i++`
Code: `print_stmt.go` `formatIncrDecr`

### F16 - Return / yield keyword spacing `implemented`
A single space between the keyword and its expression.
`return  1` → `return 1`
Code: `print_stmt.go` `formatKeywordExpr`

### F17 - If / else-if / else `implemented`
Header ends in `:`, body indented; `else if` collapses onto one line.
Code: `print_stmt.go` `formatIf`

### F18 - For loop `implemented`
`", "` between loop variables, single spaces around `in`, header ends in `:`.
`for i,x in items:` → `for i, x in items:`
Code: `print_stmt.go` `formatFor`

### F19 - While loop `implemented`
`while <cond>:`, body indented.
Code: `print_stmt.go` `formatWhile`

### F31 - Typed assignment `implemented`
A space after the type colon and around `=`. The declared type is emitted
verbatim for now (canonical `|`-union spacing is a follow-up). A trailing
`catch` block falls back to verbatim until postfix-catch is handled.
`x:int=1` → `x: int = 1`
Code: `print_stmt.go` `formatTypedAssign`

---

## Expressions & operators

### F20 - Binary operator spacing `implemented`
Single spaces around binary operators - `and`/`or`, comparisons, `in`/`not in`,
and arithmetic.
`1+2*3` → `1 + 2 * 3`
Code: `print_expr.go` `formatBinary`

### F21 - Unary operator spacing `implemented`
Word operators (`not`) take a trailing space; symbolic ones (`-`, `!`) bind
tight.
`not  c` → `not c`
Code: `print_expr.go` `formatUnary`

### F22 - Ternary spacing `implemented`
Spaces around `?` and `:`.
`cond?a:b` → `cond ? a : b`
Code: `print_expr.go` `formatTernary`

### F23 - Parentheses preserved `implemented`
Tight inside the parens; parens are never added or removed, so author grouping
is respected.
`( 1 + 2 )` → `(1 + 2)`
Code: `print_expr.go` `formatParen`

---

## Calls, paths, indexing

### F24 - Call argument spacing `implemented`
Tight parens, `", "` between arguments.
`f( a ,b )` → `f(a, b)`
Code: `print_expr.go` `formatCall`

### F32 - Named call arguments bind tight `implemented`
No spaces around `=` in a named argument - distinguishing a call-site binding
from an assignment statement.
`f(key = val)` → `f(key=val)`
Code: `print_expr.go` `formatNamedArg`

### F25 - Paths are tight `implemented`
No spaces around `.` or `[]` in a postfix chain.
`obj . method( 1 )` → `obj.method(1)`
Code: `print_expr.go` `formatPath`

### F26 - Slices are tight `implemented`
No spaces around the slice colons.
`data[ 1 : 2 ]` → `data[1:2]`
Code: `print_expr.go` `formatSlice`

---

## Collections

### F27 - List spacing `implemented`
`", "` after each element, tight brackets; empty list is `[]`.
`[ 1,2 ]` → `[1, 2]`
Code: `print_lit.go` `formatList`

### F28 - Map spacing `implemented`
Space-padded braces and a space after each key colon; `", "` between entries.
Empty map stays tight `{}`. Keys keep their original form - a bareword key is an
identifier and a quoted key is a string literal, which are semantically distinct
in Rad, so we never convert between them.
`{"x":1, y:2}` → `{ "x": 1, y: 2 }`
Code: `print_lit.go` `formatMap`

### F29 - Over-width collections and calls wrap `implemented`
When a call/list/map would exceed the target width, it breaks to one item per
line, indented one level, with a trailing comma and the closing delimiter on its
own line. A trailing comma in a flat (non-wrapped) collection is removed.
Code: `print_expr.go` `delimited`

---

## Strings & literals

### F30 - String quote normalization `deferred`
Not implemented; quotes are left exactly as written. Decided direction for when
we revisit: normalize single-quoted strings to double quotes, but only when the
swap introduces no new escapes (no raw `"` in the content); unescape now-
redundant `\'`; leave backtick, raw (`r"..."`), and triple-quoted strings alone.
Backticks in particular signal shell-command intent and may later gain inner
quotes, so they are never rewritten.

### F33 - Number / bool / null literals preserved `passthrough`
Literal value text is emitted exactly as written - no `3.140` → `3.14`
rewriting, which would be a semantic change.
Code: `print_expr.go` `formatExpr`

---

## File preamble

### F34 - Shebang preserved `passthrough`
The `#!` line is emitted verbatim.
Code: `printer.go` `verbatim`

### F35 - File header preserved `passthrough`
The `--- ... ---` header block is free text and is emitted verbatim.
Code: `printer.go` `verbatim`

---

## Roadmap

The destination is **full convergence**: any valid script, any layout, formats to
one canonical output (see the Vision in `AGENTS.md`). Everything below is emitted
verbatim today - structurally safe, just not canonicalized - and graduates to
numbered rules as we go. The list is sequenced by dependency and payoff: Tier 0
unblocks the rest, Tier 1 is the biggest single convergence win. It's a best-effort
map, not a contract - reorder as reality demands.

### Tier 0 - Engine maturity (prerequisites)
- **Comment-attachment side-table** (`DESIGN.md` Part 2). Only partially adopted
  today, which is why `F36` punts comment-bearing expressions to verbatim. Maturing
  it retires F36 and is a prerequisite for cleanly reflowing comment-heavy block
  bodies. Highest-leverage engine work.
- **`Fill` builder** (`doc.go`, currently a stub). Needed for text-flow reflow
  (long comments / strings). Lower priority - only specific rules need it.

### Tier 1 - Block constructs (the biggest coverage gap)
Each shares the `:`-header skeleton but has a bespoke body grammar; all are
verbatim today. Rough priority by how common they are in real scripts:
- **`args:`** - arg declarations, types, defaults, constraints, trailing comments.
  (Natural home for the future optional column alignment in Tier 4.)
- **`rad` / `request` / `display`** - additionally carries a **fields-first
  ordering** rule: `fields` first, then options (`transpose`/`noprint`/`quiet`/
  `sort`), then field modifiers, then `rad_if`.
- **`fn` / lambda** bodies.
- **`switch`** - case arms.
- **`command:`**.
- **`defer` / `errdefer`**.

### Tier 2 - Expression-level canonicalization
- **Long-expression wrapping** - break long boolean/arithmetic chains at operators
  with continuation indent.
- **String interpolation reflow** - reformat expressions inside `{...}`.

### Tier 3 - Content-rewriting rules (need their own value-preservation guard)
These rewrite token *content*, not just structure, so the structural guard (which
compares node kinds, not text) can't protect them - each needs a bespoke
value-preservation check plus tests. That is exactly why **F30 (string quote
normalization)** is deferred. Also here: **type-union `|` spacing** (the F31
follow-up), if/when it normalizes the declared type text.

### Tier 4 - Optional polish
- **Arg-block column alignment** - optional gofmt-style alignment of arg types and
  comments.
- **Slice spacing for complex operands** - spacing slice colons when operands are
  non-trivial, `x[a + 1 : b]` (F26 follow-up).
- **Redundant-paren removal (decision, not a given)** - today `F23` preserves
  author parens. Maximal opinionation *could* strip redundant ones, but that risks
  changing perceived intent. A call to make deliberately, not assume.

---

## Changelog / tombstones

Removed or superseded rules are recorded here; their IDs are retired, never
reused.

_(none yet)_
