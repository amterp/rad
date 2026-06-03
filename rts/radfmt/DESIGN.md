# Implementing a gofmt-style Canonical Re-printer ("rad fmt") on a Tree-sitter CST: Doc IR + Comment Attachment

> **Note:** This is the design-time research that informed the formatter, kept as
> a reference for the theory (Doc IR + comment attachment). It describes the
> *intended* design and may run ahead of, or diverge from, the current code - e.g.
> the comment-attachment side-table is only partially adopted, and `Fill` is a
> stub. For what the formatter actually does today, read the package source and
> the `doc.go` package comment.

## TL;DR
- **Doc IR:** Port Prettier's stack/worklist `printDocToString` (a productionized Wadler `best`/`be`) with a `cmds` stack of `{ind, mode, doc}` tuples, two modes `MODE_FLAT`/`MODE_BREAK`, a lookahead `fits` that measures the flat width of the next group plus the rest-of-line, a `groupModeMap` for group-ids, a buffered `lineSuffix` list flushed before every newline, and a one-time `propagateBreaks` pass so `hardline`/`break-parent` force every ancestor group to `MODE_BREAK`.
- **Comment attachment:** Treat tree-sitter `extras` as floating trivia; classify each as leading/trailing/dangling using Prettier's exact rule — `hasNewline(text, locStart, {backwards})` decides own-line vs end-of-line, then `precedingNode`/`enclosingNode`/`followingNode` (found by descending the CST by byte offset) decide attachment — and emit Doc nodes accordingly (own-line leading → `text + hardline`; same-line trailing → `lineSuffix(" " + text) + breakParent`; dangling-in-empty → `indent([hardline, text])`).
- **Blank lines:** They are NOT nodes; reconstruct them from row gaps between adjacent CST items/comments. Apply Go's exact policy: cap consecutive newlines at `maxNewlines = 2` (one blank line), preserve a single blank line, and emit none at the start/end of a block or file.

## Key Findings

**1. The rendering machine is a backtracking stack machine, not a recursive tree-walk.** Wadler's `best w k x = be k [(0,x)]` generalizes each document operation "to work on a list of indentation-document pairs" — i.e., a stack of `(indent, doc)` items. Prettier productionizes this into `printDocToString` with an explicit `cmds` stack of `{ind, mode, doc}` triples. The decision of flat vs broken is local and greedy ("locally optimal but not globally optimal"), made by the `fits` lookahead at each group.

**2. `fits` measures the flat rendering of the candidate plus the remaining command stack, stopping at the first hard line break or when the width budget goes negative.** Verbatim from Prettier 3.x's bundled printer, `fits(next, restCommands, width, hasLineSuffix, groupModeMap, mustBeFlat)` walks a local copy of the work with `restIdx = restCommands.length`; on `DOC_TYPE_STRING` it does `width -= getStringWidth(doc)`; on a `DOC_TYPE_LINE` it returns `true` if `mode === MODE_BREAK || doc.hard` (a forced break ends the line, so it fits); the loop condition is `while (width >= 0)` and the function returns `false` only when width goes negative.

**3. Breaks propagate UP the tree, statically, before rendering.** `propagateBreaks(doc)` is called once after the printer builds the doc (`docUtils.propagateBreaks(doc)` in `printAstToDoc`). `hardline` and `literalline` carry an implicit `break-parent`; any group containing a `break-parent` (transitively) has its `.break` flag set to `true`, forcing `MODE_BREAK` regardless of `fits`. Per Prettier's own docs, the printer "will try to fit everything on one line, but if it doesn't fit it will break the outermost group first and try again. It will continue breaking groups until everything fits (or there are no more groups to break)."

**4. `lineSuffix` buffers trailing comments and flushes them before any newline.** The buffer (`lineSuffix` array in the printer) accumulates docs; when a real line break is emitted, the buffer is spliced back onto the command stack ahead of the newline so the comment lands at the end of the current line. `lineSuffixBoundary` forces a flush even without a newline.

**5. Go and Prettier agree on the blank-line policy, derived from source byte/row positions.** Go's `go/printer` caps newlines at `maxNewlines = 2` (the source constant is literally `maxNewlines = 2 // max. number of newlines between source text`, lowered from 3 to 2 in an early Go change) via `nlimit(n int) int { return min(n, maxNewlines) }`. Prettier's Rationale states: "The approach that Prettier takes is to preserve empty lines the way they were in the original source code. There are two additional rules: Prettier collapses multiple blank lines into a single blank line. Empty lines at the start and end of blocks (and whole files) are removed. (Files always end with a single newline, though.)" Both reconstruct blank lines from line-number deltas, never from AST nodes.

**6. Comment classification is a two-axis decision: own-line vs end-of-line (from newlines in source), then leading/trailing/dangling (from surrounding nodes).** Prettier's `attach()` first checks `hasNewline(text, locStart(comment), {backwards:true})` (is the comment alone on its line?), then within each branch falls back to: `followingNode` → leading; `precedingNode` → trailing; `enclosingNode` → dangling; else attach to root. Go's `CommentMap` uses an equivalent line-based heuristic, but binds to the *largest* enclosing node.

## Details

### PART 1 — Doc IR and the Wadler/Prettier rendering machine

#### 1.1 The document algebra

Wadler's core algebra (from "A prettier printer") has six constructors. In his notation the laid-out form is `Doc` and the buildable form is `DOC`:

```
DOC = NIL | DOC :<> DOC | NEST Int DOC | TEXT String | LINE | DOC :<|> DOC
```

`:<>` is concatenation, `NEST i x` adds `i` to the indentation, `TEXT s` is literal text, `LINE` is a line break that becomes a space (or nothing) when flattened, and `:<|>` is the *choice* between a wide (flat) and narrow (broken) layout. `group(x)` is defined as `flatten(x) :<|> x`. Prettier expands this minimal set into a richer builder vocabulary, but every builder reduces to these primitives plus the greedy `fits` choice.

**The full node set (Prettier `commands.md`, mapped to Go):**

| Doc node | Constant (`DOC_TYPE_*`) | Meaning |
|---|---|---|
| `text(s)` | `"string"` | Literal text; must contain no newline char |
| `line` | `"line"` (`{soft:false,hard:false}`) | Becomes `" "` when flat, newline+indent when broken |
| `softline` | `"line"` `{soft:true}` | Becomes `""` when flat, newline+indent when broken |
| `hardline` | `"line"` `{hard:true}` + `break-parent` | Always a newline; forces parents to break |
| `literalline` | `"line"` `{hard:true,literal:true}` + `break-parent` | Newline with no indent; preserves trailing whitespace |
| `concat(parts)` / array | `"array"` | Sequence of docs printed in order |
| `indent(doc)` | `"indent"` | +1 indent level on the contents |
| `align(n, doc)` | `"align"` | Indent by a fixed n spaces/string |
| `group(doc, {shouldBreak,id})` | `"group"` | Try flat; if it doesn't fit, break |
| `conditionalGroup([a,b,c])` | `"group"` w/ `expandedStates` | Try each alternative least-to-most expanded |
| `fill(parts)` | `"fill"` | Break only the separators that don't fit (text-flow) |
| `ifBreak(brk, flat, {groupId})` | `"if-break"` | Print `brk` if the (referenced) group broke, else `flat` |
| `indentIfBreak(doc, {groupId})` | `"indent-if-break"` | Indent only if the referenced group broke |
| `lineSuffix(doc)` | `"line-suffix"` | Buffer doc; flush before next newline |
| `lineSuffixBoundary` | `"line-suffix-boundary"` | Force a lineSuffix flush even without a newline |
| `breakParent` | `"break-parent"` | Force all enclosing groups to break |
| `trim` | `"trim"` | Trim trailing whitespace on the current line |

#### 1.2 Wadler's `best`/`be`/`fits` (the kernel to port)

The Haskell kernel (verbatim from the paper) is:

```
best w k x = be k [(0,x)]
  be k [] = Nil
  be k ((i,NIL):z)      = be k z
  be k ((i,x :<> y):z)  = be k ((i,x):(i,y):z)
  be k ((i,NEST j x):z) = be k ((i+j,x):z)
  be k ((i,TEXT s):z)   = Text s (be (k+length s) z)
  be k ((i,LINE):z)     = Line i (be i z)
  be k ((i,x :<|> y):z) = better k (be k ((i,x):z)) (be k ((i,y):z))
  better k x y = if fits (w-k) x then x else y

fits w x | w < 0 = False
fits w Nil               = True
fits w (Text s x)        = fits (w - length s) x
fits w (Line i x)        = True
```

Two behavioral subtleties (made explicit in Hodgson's imperative port) that you MUST preserve:
- **`fits` measures `x` *and* the rest of the stack `z`, not just `x`.** Choosing flat for `x` can still overflow later in the same line, so the choice operator concatenates the candidate onto the remaining document before measuring. This is why Prettier's `fits` takes `restCommands`.
- **`fits` only reads to the end of the current line** (it returns `True` on the first `Line`/break), bounding lookahead to ~`printWidth` characters — in Haskell via laziness, in an imperative port by returning early on a hard/break-mode line. This keeps the algorithm O(n).

#### 1.3 Prettier's `printDocToString` main loop (verbatim structure)

The modern engine replaces Wadler's two layouts with two **modes** (`MODE_BREAK`, `MODE_FLAT`) carried per stack frame, plus a `groupModeMap` keyed by group-id, a `pos` column counter, and a `lineSuffix` buffer. The loop pops `{ind, mode, doc}` from `cmds` and dispatches on `getDocType(doc)`:

- `DOC_TYPE_STRING`: `out.push(doc); pos += getStringWidth(doc)`.
- `DOC_TYPE_ARRAY`: push parts in reverse so they pop in order: `for (let i = doc.length - 1; i >= 0; i--) cmds.push({ind, mode, doc: doc[i]})`.
- `DOC_TYPE_INDENT`: `cmds.push({ind: makeIndent(ind, options), mode, doc: doc.contents})`.
- `DOC_TYPE_ALIGN`: `makeAlign(ind, doc.n, options)`.
- `DOC_TYPE_TRIM`: `pos -= trim(out)` (strip trailing whitespace already emitted).
- `DOC_TYPE_GROUP`: see §1.4 below.
- `DOC_TYPE_IF_BREAK` / `DOC_TYPE_INDENT_IF_BREAK`: resolve the controlling mode — `const groupMode = doc.groupId ? groupModeMap[doc.groupId] || MODE_FLAT : mode;` then choose `doc.breakContents` (if `MODE_BREAK`) or `doc.flatContents`.
- `DOC_TYPE_LINE`: if `mode === MODE_BREAK` (or `doc.hard`): flush lineSuffix if present, then emit newline + indentation (`pos = ind.length`); else if `!doc.soft`, emit `" "` and `pos++`.
- `DOC_TYPE_LINE_SUFFIX`: `lineSuffix.push({ind, mode, doc: doc.contents})` — buffer, don't print.
- `DOC_TYPE_LINE_SUFFIX_BOUNDARY`: if `lineSuffix.length`, inject a synthetic hardline frame to force a flush.
- `DOC_TYPE_BREAK_PARENT`: no-op at print time (already consumed by `propagateBreaks`).

After the main pop, if the frame's doc had an `id`, record `groupModeMap[doc.id] = <the mode it resolved to>` so later `ifBreak(..., {groupId})` can reference it.

**lineSuffix flush rule:** when a newline is about to be emitted (a `DOC_TYPE_LINE` in `MODE_BREAK`/hard) and `lineSuffix.length > 0`, the printer pushes the current line doc back and splices the buffered suffix frames ahead of it, so trailing comments print at the end of the line *before* the break. At end-of-document, any remaining lineSuffix is flushed.

#### 1.4 How `group` chooses flat vs broken

In the `DOC_TYPE_GROUP` case the printer branches on the current `mode`:

- **`MODE_FLAT` and `!shouldRemeasure`**: stay flat — `cmds.push({ind, mode: doc.break ? MODE_BREAK : MODE_FLAT, doc: doc.contents})`.
- **`MODE_BREAK`** (or remeasure): compute `rem = width - pos` and `hasLineSuffix = lineSuffix.length > 0`, build a flat candidate `next = {ind, mode: MODE_FLAT, doc: doc.contents}`, and test:
  ```js
  if (!doc.break && fits(next, cmds, rem, hasLineSuffix, groupModeMap)) {
    cmds.push(next);                 // group fits flat
  } else {
    cmds.push({ind, mode: MODE_BREAK, doc: doc.contents}); // break it
  }
  ```
  If `doc.break` is set (via `propagateBreaks` or `shouldBreak`), the `fits` test is skipped and the group goes straight to `MODE_BREAK`. Because the printer breaks the **outermost** group first and re-tests inner groups, breaking cascades from the outside in until everything fits or there are no more groups to break.

**Conditional groups / `expandedStates`:** when a group carries `expandedStates` (from `conditionalGroup([a, b, c])`), the printer tries each state least-expanded first: if `doc.break`, it jumps to the most-expanded (`expandedStates[last]`); otherwise it loops `for (let i = 1; i < expandedStates.length + 1; i++)`, testing `fits` on `expandedStates[i]` in `MODE_FLAT`, taking the first that fits, and falling back to the most-expanded state if none do. Prettier's `commands.md` documents the exact contract and the cost: "This should be used as last resort as it triggers an exponential complexity when nested… This will try to print the first alternative, if it fit use it, otherwise go to the next one and so on. The alternatives is an array of documents going from the least expanded (most flattened) representation first to the most expanded." Use sparingly.

#### 1.5 `fits` — verbatim and annotated (Prettier 3.x)

```js
function fits(next, restCommands, width, hasLineSuffix, groupModeMap, mustBeFlat) {
  let restIdx = restCommands.length;
  const cmds = [next];
  const out = [];
  while (width >= 0) {
    if (cmds.length === 0) {
      if (restIdx === 0) return true;        // consumed whole rest-of-line: it fits
      cmds.push(restCommands[--restIdx]);    // pull next rest-command (backwards)
      continue;
    }
    const { mode, doc } = cmds.pop();
    switch (getDocType(doc)) {
      case DOC_TYPE_STRING:
        out.push(doc); width -= getStringWidth(doc); break;
      case DOC_TYPE_ARRAY: case DOC_TYPE_FILL: {
        const parts = getDocParts(doc);
        for (let i = parts.length - 1; i >= 0; i--) cmds.push({ mode, doc: parts[i] });
        break;
      }
      case DOC_TYPE_INDENT: case DOC_TYPE_ALIGN:
      case DOC_TYPE_INDENT_IF_BREAK: case DOC_TYPE_LABEL:
        cmds.push({ mode, doc: doc.contents }); break;
      case DOC_TYPE_TRIM: width += trim(out); break;
      case DOC_TYPE_GROUP: {
        if (mustBeFlat && doc.break) return false;
        const groupMode = doc.break ? MODE_BREAK : mode;
        const contents = doc.expandedStates && groupMode === MODE_BREAK
          ? doc.expandedStates.at(-1) : doc.contents;
        cmds.push({ mode: groupMode, doc: contents }); break;
      }
      case DOC_TYPE_IF_BREAK: {
        const groupMode = doc.groupId ? groupModeMap[doc.groupId] || MODE_FLAT : mode;
        const contents = groupMode === MODE_BREAK ? doc.breakContents : doc.flatContents;
        if (contents) cmds.push({ mode, doc: contents });
        break;
      }
      case DOC_TYPE_LINE:
        if (mode === MODE_BREAK || doc.hard) return true;  // line ends here -> fits
        if (!doc.soft) { out.push(" "); width--; }
        break;
      case DOC_TYPE_LINE_SUFFIX: hasLineSuffix = true; break;
      case DOC_TYPE_LINE_SUFFIX_BOUNDARY: if (hasLineSuffix) return false; break;
    }
  }
  return false;     // width went negative
}
```

Key semantics: `fits` returns **true** when it consumes the candidate and the entire rest-of-line without exhausting the width budget, OR it hits a line break that would end the current line (`MODE_BREAK` line or `doc.hard`). It returns **false** when `width` goes negative, when a `mustBeFlat` candidate contains a forced-break group, or when a `lineSuffixBoundary` is reached with a pending line-suffix. (Note: across versions the signature has a minor variation — current `main` inserts an `options` parameter: `fits(next, restCommands, width, options, hasLineSuffix, groupModeMap, mustBeFlat)`; the body logic is identical.)

#### 1.6 `propagateBreaks` (break-parent propagation)

Before printing, traverse the doc bottom-up. Any `break-parent` (and the implicit one inside every `hardline`/`literalline`) marks its nearest enclosing group `.break = true`; that propagates outward so **every** ancestor group breaks. This is a static, one-pass analysis — "this only matters for 'hard' breaks … that can be statically analyzed." A `group` containing a deeply nested `hardline` therefore always renders broken; this is the mechanism behind Prettier's rule that "Functions always break after the opening curly brace no matter what, so the array breaks as well for consistent formatting" — a multiline function body's hardline forces every enclosing call/array group open.

#### 1.7 Worked example A — function call argument list

Doc construction (the canonical Prettier `ArrayExpression`/call shape):

```
group([
  "foo(",
  indent([ softline, join([",", line], args) ]),
  softline,
  ")"
])
```

With `args = [reallyLongArg(), omgSoManyParameters(), IShouldRefactorThis(), isThereSeriouslyAnotherOne()]` at width 100:

- **Flat (fits):** `softline`→`""`, `line`→`" "`:
  `foo(reallyLongArg(), omgSoManyParameters(), IShouldRefactorThis(), isThereSeriouslyAnotherOne())`
- **Broken (exceeds 100, group → MODE_BREAK):** each `softline`/`line` becomes a newline+indent:
  ```
  foo(
    reallyLongArg(),
    omgSoManyParameters(),
    IShouldRefactorThis(),
    isThereSeriouslyAnotherOne(),
  )
  ```
The trailing comma in broken mode is produced with `ifBreak(",")` after the last element. The decision is made once, at the group, by `fits(flatCandidate, restCommands, 100 - pos, …)`.

#### 1.8 Worked example B — trailing comment surviving a wrap

`["a", lineSuffix(" // comment"), ";", hardline]` renders as:

```
a; // comment
```

The `lineSuffix(" // comment")` is buffered when first popped; `";"` prints normally; when the `hardline` is reached the buffer is flushed *before* the newline, so the comment lands at end-of-line after the semicolon — never inside the code. If the line wraps, the same flush-before-newline rule guarantees the comment stays attached to the visual end of the construct's line, and the `break-parent` inside `hardline` forces the enclosing group broken.

#### 1.9 Go-flavored Doc IR types and render worklist

```go
type Mode uint8
const ( ModeFlat Mode = iota; ModeBreak )

// Doc is the sealed interface implemented by every node kind.
type Doc interface{ isDoc() }

type Text  struct{ S string }                       // no '\n' allowed
type Line  struct{ Soft, Hard, Literal bool }        // hard/literal carry implicit break-parent
type Concat struct{ Parts []Doc }
type Indent struct{ Contents Doc }
type Align  struct{ N int; Contents Doc }
type Group  struct {
    Contents       Doc
    Break          bool      // set by propagateBreaks or shouldBreak
    ID             GroupID   // 0 = none
    ExpandedStates []Doc      // conditionalGroup; nil if plain group
}
type Fill          struct{ Parts []Doc }
type IfBreak       struct{ BreakContents, FlatContents Doc; GroupID GroupID }
type IndentIfBreak struct{ Contents Doc; GroupID GroupID }
type LineSuffix    struct{ Contents Doc }
type LineSuffixBoundary struct{}
type BreakParent   struct{}
type Trim          struct{}

func (Text) isDoc() {}; func (Line) isDoc() {} /* …all kinds… */

type GroupID uint32

type cmd struct{ ind Indentation; mode Mode; doc Doc }
type Indentation struct{ value string }   // accumulated indent string

func PrintDocToString(doc Doc, width int) string {
    groupModeMap := map[GroupID]Mode{}
    var out []string
    pos := 0
    cmds := []cmd{{Indentation{}, ModeBreak, doc}}
    var lineSuffix []cmd
    shouldRemeasure := false

    for len(cmds) > 0 {
        c := cmds[len(cmds)-1]; cmds = cmds[:len(cmds)-1]
        switch d := c.doc.(type) {

        case Text:
            out = append(out, d.S); pos += stringWidth(d.S)

        case Concat:
            for i := len(d.Parts) - 1; i >= 0; i-- {
                cmds = append(cmds, cmd{c.ind, c.mode, d.Parts[i]})
            }
        case Indent:
            cmds = append(cmds, cmd{makeIndent(c.ind), c.mode, d.Contents})
        case Align:
            cmds = append(cmds, cmd{makeAlign(c.ind, d.N), c.mode, d.Contents})
        case Trim:
            pos -= trimTrailing(&out)

        case Group:
            switch c.mode {
            case ModeFlat:
                if !shouldRemeasure {
                    m := ModeFlat; if d.Break { m = ModeBreak }
                    cmds = append(cmds, cmd{c.ind, m, d.Contents})
                    break
                }
                fallthrough
            case ModeBreak:
                shouldRemeasure = false
                next := cmd{c.ind, ModeFlat, d.Contents}
                rem := width - pos
                hasLS := len(lineSuffix) > 0
                if !d.Break && fits(next, cmds, rem, hasLS, groupModeMap, false) {
                    cmds = append(cmds, next)
                } else if d.ExpandedStates != nil {
                    cmds = append(cmds, chooseExpanded(d, c.ind, cmds, rem, hasLS, groupModeMap))
                } else {
                    cmds = append(cmds, cmd{c.ind, ModeBreak, d.Contents})
                }
            }
            if d.ID != 0 {
                groupModeMap[d.ID] = cmds[len(cmds)-1].mode
            }

        case IfBreak:
            gm := c.mode
            if d.GroupID != 0 { gm = groupModeMap[d.GroupID] } // default ModeFlat (zero value)
            chosen := d.FlatContents
            if gm == ModeBreak { chosen = d.BreakContents }
            if chosen != nil { cmds = append(cmds, cmd{c.ind, c.mode, chosen}) }

        case IndentIfBreak:
            gm := groupModeMap[d.GroupID]
            inner := d.Contents
            if gm == ModeBreak { inner = Indent{d.Contents} }
            cmds = append(cmds, cmd{c.ind, c.mode, inner})

        case LineSuffix:
            lineSuffix = append(lineSuffix, cmd{c.ind, c.mode, d.Contents})

        case LineSuffixBoundary:
            if len(lineSuffix) > 0 {
                cmds = append(cmds, cmd{c.ind, c.mode, Line{Hard: true}})
            }

        case BreakParent: // no-op at print time

        case Line:
            if c.mode == ModeFlat && !d.Hard {
                if !d.Soft { out = append(out, " "); pos++ }
                break
            }
            // MODE_BREAK or hard line: flush buffered trailing comments first
            if len(lineSuffix) > 0 {
                cmds = append(cmds, c)                 // re-process this Line after suffix
                for i := len(lineSuffix) - 1; i >= 0; i-- {
                    cmds = append(cmds, lineSuffix[i])
                }
                lineSuffix = nil
                break
            }
            if d.Literal {
                out = append(out, "\n"); pos = 0
            } else {
                trimTrailing(&out)
                out = append(out, "\n"+c.ind.value); pos = len(c.ind.value)
            }
        }
    }
    // flush any trailing line-suffix at EOF
    for _, ls := range lineSuffix { /* append rendered ls */ _ = ls }
    return strings.Join(out, "")
}

func fits(next cmd, rest []cmd, width int, hasLineSuffix bool,
          gmm map[GroupID]Mode, mustBeFlat bool) bool {
    restIdx := len(rest)
    cmds := []cmd{next}
    var out []string
    for width >= 0 {
        if len(cmds) == 0 {
            if restIdx == 0 { return true }
            restIdx--; cmds = append(cmds, rest[restIdx]); continue
        }
        c := cmds[len(cmds)-1]; cmds = cmds[:len(cmds)-1]
        switch d := c.doc.(type) {
        case Text:
            out = append(out, d.S); width -= stringWidth(d.S)
        case Concat:
            for i := len(d.Parts) - 1; i >= 0; i-- {
                cmds = append(cmds, cmd{c.ind, c.mode, d.Parts[i]})
            }
        case Fill:
            for i := len(d.Parts) - 1; i >= 0; i-- {
                cmds = append(cmds, cmd{c.ind, c.mode, d.Parts[i]})
            }
        case Indent:        cmds = append(cmds, cmd{c.ind, c.mode, d.Contents})
        case Align:         cmds = append(cmds, cmd{c.ind, c.mode, d.Contents})
        case IndentIfBreak: cmds = append(cmds, cmd{c.ind, c.mode, d.Contents})
        case Trim:          width += /* trim */ 0
        case Group:
            if mustBeFlat && d.Break { return false }
            gm := c.mode; if d.Break { gm = ModeBreak }
            contents := d.Contents
            if d.ExpandedStates != nil && gm == ModeBreak {
                contents = d.ExpandedStates[len(d.ExpandedStates)-1]
            }
            cmds = append(cmds, cmd{c.ind, gm, contents})
        case IfBreak:
            gm := c.mode
            if d.GroupID != 0 { gm = gmm[d.GroupID] }
            chosen := d.FlatContents
            if gm == ModeBreak { chosen = d.BreakContents }
            if chosen != nil { cmds = append(cmds, cmd{c.ind, c.mode, chosen}) }
        case Line:
            if c.mode == ModeBreak || d.Hard { return true }
            if !d.Soft { out = append(out, " "); width-- }
        case LineSuffix:          hasLineSuffix = true
        case LineSuffixBoundary:  if hasLineSuffix { return false }
        }
    }
    return false
}
```

`propagateBreaks` runs once before `PrintDocToString`:

```go
func propagateBreaks(d Doc) (containsForcedBreak bool) {
    switch n := d.(type) {
    case BreakParent: return true
    case Line:        return n.Hard       // hardline/literalline force breaks
    case *Group:
        child := propagateBreaks(n.Contents)
        for _, st := range n.ExpandedStates { if propagateBreaks(st) { child = true } }
        if child { n.Break = true }
        return n.Break
    case Concat:
        forced := false
        for _, p := range n.Parts { if propagateBreaks(p) { forced = true } }
        return forced
    /* …recurse into Indent/Align/IfBreak/LineSuffix contents… */
    }
    return false
}
```
(Use pointer receivers for `Group` so the `.Break` mutation persists.)

---

### PART 2 — Comment & blank-line attachment

#### 2.1 Reference approaches

**Prettier (`src/main/comments/attach.js`).** Comments are decorated then attached. `decorateComment(node, comment, text)` finds, by binary search over `getSortedChildNodes(node)` (children sorted by `[locStart, locEnd]`), the child that encloses the comment (`locStart(child) <= locStart(comment) && locEnd(comment) <= locEnd(child)`) and recurses into it; the deepest such node becomes `enclosingNode`. Among the enclosing node's children, the nearest child ending at/before the comment is `precedingNode`, and the nearest child starting at/after it is `followingNode`. Then `attach()` dispatches on three cases:

1. **`hasNewline(text, locStart(comment), {backwards:true})`** — the comment starts its own line (own-line): after language-specific handlers, fall back to `followingNode ? addLeadingComment : precedingNode ? addTrailingComment : enclosingNode ? addDanglingComment : addDanglingComment(ast)`.
2. **else if `hasNewline(text, locEnd(comment))`** — there's code before but not after the comment on its line (end-of-line): fall back to `precedingNode ? addTrailingComment : followingNode ? addLeadingComment : addDanglingComment`.
3. **else** — the comment is wedged between code on both sides (remaining): prefer attaching as trailing/leading per language handlers.

Printing (`printComments`): for each attached comment, `leading` comments print before the node and, if `hasNewline` after the comment, add a `hardline`; `trailing` comments print via `printTrailingComment` (using `lineSuffix` for same-line ones); dangling comments print via `printDanglingComments`, which returns `indent([hardline, join(hardline, parts)])` (or `join(hardline, parts)` when `sameIndent`).

**Go (`go/ast/commentmap.go` + `go/printer/printer.go`).** `NewCommentMap` associates a `*CommentGroup` `g` with node `n` if: (1) `g` starts on the same line as `n` ends; (2) `g` starts on the line immediately following `n` AND there's an empty line after `g` before the next node; or (3) `g` starts before `n` and isn't already associated with the previous node. It associates to the **largest** node possible (a trailing line-comment binds to the whole assignment, not the last operand). The printer's `intersperseComments` writes pending comments before the next token; it then ensures a line break "after a //-style comment, before EOF, and before a closing '}'", and for a `/*…*/` comment followed on the same line by a non-comma/non-closing token it inserts a single separating space.

**Biome (CST trivia model).** Biome attaches trivia directly to tokens. Its architecture docs state the rule verbatim: "Every trivia up to the token/keyword (including line breaks) will be the leading trivia; everything until the next linebreak (but not including it) will be the trailing trivia." In their worked example, `// comment 1` is trailing trivia of the `;` token and `// comment 2` is leading trivia of the next `const` keyword (the CST shows `CONST_KW@27..45 "const" [Newline("\n"), Comments("// comment 2"), Newline("\n")]`). This is a purely position/newline-driven split — the same two-axis rule as Prettier, but resolved at the token level rather than by node search.

#### 2.2 Concrete algorithm for rad fmt (tree-sitter CST, Go, width 100)

**Inputs:** the CST root, the original `source []byte`, and a flat list of comment nodes (the tree-sitter `extras`, gathered by walking the tree and collecting nodes whose `IsExtra()` / kind is `comment`). Each node exposes `StartByte/EndByte` and `StartPoint/EndPoint` (row, column).

**Step 0 — gather and sort.** Collect all extra/comment nodes; sort by `StartByte`. Collect the "real" (non-extra, named) nodes for sibling/offset queries.

**Step 1 — find enclosing/preceding/following for each comment** (port of `decorateComment`):
```
func locate(root, comment):
    enclosing = root
    descend:
      children = namedChildren(enclosing) sorted by StartByte   // skip extras
      pick child where child.StartByte <= comment.StartByte
                    && comment.EndByte <= child.EndByte
      if found: enclosing = child; repeat descend
      else: break
    preceding = last child of enclosing with child.EndByte <= comment.StartByte
    following = first child of enclosing with child.StartByte >= comment.EndByte
    return preceding, enclosing, following
```
Use byte offsets for containment (robust against multibyte columns); use **row numbers** (`StartPoint.Row`) for the newline tests below.

**Step 2 — classify own-line vs end-of-line trailing.** Let `prevTok` be the last token/node ending before the comment and `nextTok` the first starting after it.
- **Own-line** iff there is a newline between `prevTok.EndPoint.Row` and `comment.StartPoint.Row` (i.e., `comment.StartPoint.Row > prevTok.EndPoint.Row`), or the comment is the first thing in the file/block. Equivalent to Prettier's `hasNewline(text, locStart, {backwards:true})`.
- **End-of-line (same-line trailing)** iff `comment.StartPoint.Row == prevTok.EndPoint.Row` (code precedes it on the line) AND a newline follows (`nextTok.StartPoint.Row > comment.EndPoint.Row`). Equivalent to `hasNewline(text, locEnd)`.
- **Remaining** (rare): code on both sides, same line — treat as trailing of `preceding`.

**Step 3 — leading/trailing/dangling decision** (port of `attach` fallbacks):
- Own-line: `following ? leading(following) : preceding ? trailing(preceding) : dangling(enclosing)`.
- End-of-line: `preceding ? trailing(preceding) : following ? leading(following) : dangling(enclosing)`.
- If neither `preceding` nor `following` exists (empty container), always `dangling(enclosing)`.

**Step 4 — emit Doc nodes per situation** (the mapping table in §2.3).

**Step 5 — blank-line reconstruction** (port of Go's `nlimit`/`maxNewlines = 2` and Prettier's collapse rule). Blank lines are NOT CST nodes; derive them from row gaps. For any two adjacent emitted items `A` then `B` (statements, decls, or comments) inside a block:
```
gap   = B.StartPoint.Row - A.EndPoint.Row   // rows strictly between them
blanks = gap - 1                            // number of empty source lines
// normal separator between block items is one hardline;
// if blanks >= 1, emit a SECOND hardline (one blank line)
```
Concretely: the normal separator between block items is a single `hardline`; if `blanks >= 1` (≥1 empty source line between them), emit `hardline, hardline` (one blank line). Never emit more than one blank line — this is exactly Go's cap, whose source constant is `maxNewlines = 2 // max. number of newlines between source text` enforced by `func nlimit(n int) int { return min(n, maxNewlines) }`. Suppress the leading separator before the first item and the trailing separator after the last item of a block/file; Prettier's Rationale states the equivalent rule: "Empty lines at the start and end of blocks (and whole files) are removed. (Files always end with a single newline, though.)" Prettier implements blank-line preservation with `isNextLineEmpty`/`isPreviousLineEmpty` over the original text.

**`isNextLineEmpty(source, node)`** for rad fmt: starting at `node.EndByte`, skip spaces/tabs and one inline/trailing comment, skip one newline, then check whether the next char (after more spaces) is another newline — if so, a blank line follows the node.

#### 2.3 Edge-case table (situation → chosen model → Doc nodes)

| # | Edge case | How the model handles it | Doc nodes to emit |
|---|---|---|---|
| 1 | Comment between block-opening `{` and first child | `preceding = {` token, `following =` first stmt, comment on its own line → **leading** of first child | `text(comment) + hardline` prepended to first child; if a blank line followed in source, `hardline + hardline` |
| 2 | Trailing comment on a construct that line-wraps | End-of-line trailing of `preceding`; `lineSuffix` guarantees flush-before-newline so it can't fall inside wrapped code; `hardline`'s `break-parent` keeps the wrapped group broken | `lineSuffix(" " + text) + breakParent` |
| 3 | Comment at EOF (after last statement) | `following = nil`, `preceding =` last stmt → **trailing** of last stmt (or **dangling** of root if no preceding). Go forces a newline before EOF | own-line → `hardline + text`; trailing → `lineSuffix(" " + text)`; always end file with single `hardline` |
| 4 | Leading file/license comment above first statement | Own-line, `following =` first decl → **leading**; preserve the blank line gap to the first decl | `text + hardline (+ hardline if blank-line gap)` then the decl |
| 5 | Dangling comment in empty block/container `{}` / `()` | No `preceding`/`following` inside the container → **dangling** of `enclosing`; container must still render its delimiters | `group(["{", indent([hardline, join(hardline, comments)]), hardline, "}"])` → i.e. `indent([hardline, text])` + closing `hardline` |
| 6 | Comment between elements of a delimited list | If own-line → **leading** of the following element (prints above it inside the indented list); if same-line after a `,` → **trailing** of the preceding element | own-line: `text + hardline` before next element; trailing: `lineSuffix(" " + text)` after the element's comma |

#### 2.4 Doc-node mapping summary

- **Leading own-line comment** → `text(comment)`, then `hardline` (plus a second `hardline` if a blank line separated it from the node). The `hardline` propagates `break-parent` to enclosing groups, which is correct: a node with an own-line leading comment can't stay flat.
- **Leading same-line comment** (block `/* */` before code on same line) → `text(comment) + " "` (a space, no break).
- **Trailing same-line comment** → `lineSuffix(" " + text)` (+ implicit `breakParent` if it's a `//` line comment, since the rest of the line must end). For Go, every `//` comment forces a following newline.
- **Trailing own-line comment** (comment owned as trailing but on its own line below) → `lineSuffix([hardline, text])`.
- **Dangling comment (empty block)** → `indent([hardline, text])` before the closing delimiter's `hardline`.
- **Blank line** → a second `hardline` (a doubled `hardline`), capped so 2+ source blanks collapse to one.

## Recommendations

**Stage 1 — build and validate the Doc engine first (in isolation).** Implement the IR types and `PrintDocToString`/`fits`/`propagateBreaks` from §1.9 exactly as Prettier's, with a hand-written `cmds` stack (no recursion). Validate against the two worked examples (call-arg wrapping at width 100; trailing-comment lineSuffix) and a property test: **idempotence** — `format(format(x)) == format(x)`. This is the single most important correctness invariant for a canonical re-printer and the one Prettier's own `--debug-check` enforces. *Benchmark to advance:* idempotent on a corpus of ≥1000 real Go files and byte-identical to a reference on round-tripping already-formatted code.

**Stage 2 — implement comment attachment as a separate pass over the CST**, producing per-node `leadingComments`/`trailingComments`/`danglingComments` slices (mutating a side-table keyed by node ID, not the immutable CST). Port `locate()` (§2.2 step 1) and the two-axis classifier (steps 2–3) verbatim from Prettier's `attach`. Add the language-specific handlers you need **only** for the edge cases in §2.3 — start with cases 1 (open-brace), 5 (empty block), and 6 (list elements), which cause the most visible misplacement. *Benchmark:* every comment in the corpus is emitted exactly once (port Prettier's `ensureAllCommentsPrinted` assertion — a dropped comment is a hard failure) and lands on the same logical line as in gofmt's output.

**Stage 3 — blank-line reconstruction.** Implement `isNextLineEmpty`/`isPreviousLineEmpty` over the source bytes and the `gap-1`/`maxNewlines=2` collapse from §2.2 step 5. Wire the doubled-`hardline` separator into block/statement-sequence printers, suppressing it at block start/end. *Benchmark:* blank-line placement byte-identical to gofmt on the corpus.

**Thresholds that change the plan:**
- If idempotence fails, **stop and fix the Doc engine** before touching comments — non-idempotence almost always traces to `fits` measuring the wrong rest-of-line or a missing `propagateBreaks`.
- If comments land on the wrong node, prefer Go's "attach to the largest node" rule over Prettier's deepest-enclosing rule for statement-level comments (it matches gofmt expectations and avoids the gosec-style `#nosec` misattachment bug where a comment binds to a sub-expression instead of the whole `if`).
- If `conditionalGroup`/`expandedStates` causes blowups, replace with a plain `group` + `ifBreak`; reserve conditional groups for the few constructs that truly need 3+ layouts.

## Caveats

- **Source/version drift in `fits`.** The verbatim `fits` above is from Prettier 3.0.0-alpha.1's bundled `doc.mjs`; current `main` inserts an `options` parameter (`fits(next, restCommands, width, options, hasLineSuffix, groupModeMap, mustBeFlat)`) but the body logic is unchanged. The un-bundled `src/document/printer.js` uses the same logic with names `getDocType`/`getStringWidth`/`getDocParts`/`trim`/`at`. GitHub raw/blob fetches were blocked during research, so exact upstream line numbers for the un-bundled file are not quoted.
- **CST vs AST.** Prettier and Go both attach comments to an **AST**; you are working on a tree-sitter **CST** where comments are `extras` that appear as un-fielded siblings anywhere. This is actually *easier* for attachment (comments are already in the tree in source order) but means you must (a) skip `extras` when computing `namedChildren` for `locate()`, and (b) reconstruct blank lines from `StartPoint.Row`/`EndByte` because tree-sitter discards whitespace. Use byte offsets for containment and row numbers for newline tests.
- **Go's "largest node" heuristic is deliberately different from Prettier's "deepest enclosing."** They produce different attachments for trailing comments on compound statements. For a gofmt-*style* tool, the Go rule is the safer default; adopt it explicitly rather than copying Prettier's deepest-node search wholesale.
- **`go/printer` is not Wadler-based.** Go's printer is a single-pass tabwriter-backed emitter, not a Doc/`fits` engine; it never does width-based group breaking the way Prettier does. Use it as the authority for **comment interspersing and blank-line policy** (`intersperseComments`, `nlimit`, `maxNewlines`), and use Prettier as the authority for the **Doc IR and line-breaking**. Mixing the two is the core design of rad fmt.
- **`printWidth` is a target, not a hard cap.** Per Prettier's own docs, the engine "will make both shorter and longer lines" — e.g. a long unbreakable string or a `//` comment can exceed 100 columns. Don't assert a hard 100-column maximum in tests; assert idempotence and structural correctness instead.
- **Known hard cases the references still get wrong**, which you should expect and test: multiple comments on one line between two statements (Prettier PR #9672 had to fix reordering), comments around `else`/`else if` and ternary branches, and comments wedged between a selector expression and its `.` — documented in golang/go issue #70978 ("go/printer: Comment-LineBreak-LineBreak-SelectorExpr-Comment AST modification issue", reported on go1.23.4), where lengthening the selector's operand makes go/printer intersperse a following comment *in the middle of* the selector expression. Handle these with explicit per-construct handlers rather than relying on the generic fallback.