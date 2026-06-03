package radfmt

import (
	"strings"
	"unicode/utf8"
)

// stringWidth approximates the rendered column width of s. It counts runes
// rather than bytes so multibyte UTF-8 doesn't overcount. (Wide/zero-width
// glyphs aren't special-cased yet; width is a target, not a hard cap.)
func stringWidth(s string) int { return utf8.RuneCountInString(s) }

// MaxWidth is the target line width. It is a target, not a hard cap: long
// unbreakable tokens (string literals, comments) may exceed it. Tests assert
// idempotence and structural equivalence, never a hard column maximum.
const MaxWidth = 100

// IndentUnit is one level of indentation.
const IndentUnit = "    " // 4 spaces

type mode uint8

const (
	modeFlat mode = iota
	modeBreak
)

// cmd is one frame on the render worklist: a doc to print at a given
// indentation in a given mode.
type cmd struct {
	ind  string
	mode mode
	doc  Doc
}

// PrintDocToString renders a Doc to its final string at the given target width.
// It is a stack machine (no recursion), a direct port of Prettier's
// printDocToString: pop a frame, dispatch on doc kind, push children. Group
// nodes consult fits() to choose flat vs broken. propagateBreaks must run on doc
// before this is called.
func PrintDocToString(doc Doc, width int) string {
	propagateBreaks(doc)

	groupModeMap := map[GroupID]mode{}
	var out []string
	pos := 0
	cmds := []cmd{{ind: "", mode: modeBreak, doc: doc}}
	var lineSuffixes []cmd
	shouldRemeasure := false

	for len(cmds) > 0 {
		c := cmds[len(cmds)-1]
		cmds = cmds[:len(cmds)-1]

		switch d := c.doc.(type) {
		case Text:
			out = append(out, d.S)
			pos += stringWidth(d.S)

		case Concat:
			for i := len(d.Parts) - 1; i >= 0; i-- {
				cmds = append(cmds, cmd{c.ind, c.mode, d.Parts[i]})
			}

		case Fill:
			// Simplified fill: treat like concat for now. Proper fill (per-gap
			// breaking) lands with width-aware collection wrapping.
			for i := len(d.Parts) - 1; i >= 0; i-- {
				cmds = append(cmds, cmd{c.ind, c.mode, d.Parts[i]})
			}

		case Indent:
			cmds = append(cmds, cmd{c.ind + IndentUnit, c.mode, d.Contents})

		case Align:
			cmds = append(cmds, cmd{c.ind + strings.Repeat(" ", d.N), c.mode, d.Contents})

		case Trim:
			pos -= trimTrailing(&out)

		case *Group:
			switch c.mode {
			case modeFlat:
				if !shouldRemeasure {
					m := modeFlat
					if d.Break {
						m = modeBreak
					}
					cmds = append(cmds, cmd{c.ind, m, d.Contents})
					if d.ID != 0 {
						groupModeMap[d.ID] = m
					}
					break
				}
				fallthrough
			case modeBreak:
				shouldRemeasure = false
				next := cmd{c.ind, modeFlat, d.Contents}
				rem := width - pos
				hasLS := len(lineSuffixes) > 0
				if !d.Break && fits(next, cmds, rem, hasLS, groupModeMap, false) {
					cmds = append(cmds, next)
					if d.ID != 0 {
						groupModeMap[d.ID] = modeFlat
					}
				} else if d.ExpandedStates != nil {
					chosen := chooseExpanded(d, c.ind, cmds, rem, hasLS, groupModeMap)
					cmds = append(cmds, chosen)
					if d.ID != 0 {
						groupModeMap[d.ID] = chosen.mode
					}
				} else {
					cmds = append(cmds, cmd{c.ind, modeBreak, d.Contents})
					if d.ID != 0 {
						groupModeMap[d.ID] = modeBreak
					}
				}
			}

		case IfBreak:
			gm := c.mode
			if d.GroupID != 0 {
				gm = groupModeMap[d.GroupID] // zero value modeFlat if unseen
			}
			chosen := d.FlatContents
			if gm == modeBreak {
				chosen = d.BreakContents
			}
			if chosen != nil {
				cmds = append(cmds, cmd{c.ind, c.mode, chosen})
			}

		case IndentIfBreak:
			gm := groupModeMap[d.GroupID]
			inner := d.Contents
			if gm == modeBreak {
				inner = Indent{Contents: d.Contents}
			}
			cmds = append(cmds, cmd{c.ind, c.mode, inner})

		case LineSuffix:
			lineSuffixes = append(lineSuffixes, cmd{c.ind, c.mode, d.Contents})

		case LineSuffixBoundary:
			if len(lineSuffixes) > 0 {
				cmds = append(cmds, cmd{c.ind, c.mode, Line{Hard: true}})
			}

		case BreakParent:
			// no-op at print time; consumed by propagateBreaks

		case Line:
			if c.mode == modeFlat && !d.Hard {
				if !d.Soft {
					out = append(out, " ")
					pos++
				}
				break
			}
			// Break mode or hard line: flush buffered line suffixes first, so
			// trailing comments land before the newline.
			if len(lineSuffixes) > 0 {
				cmds = append(cmds, c) // re-process this Line after the suffixes
				for i := len(lineSuffixes) - 1; i >= 0; i-- {
					cmds = append(cmds, lineSuffixes[i])
				}
				lineSuffixes = nil
				break
			}
			if d.Literal {
				out = append(out, "\n")
				pos = 0
			} else {
				trimTrailing(&out)
				out = append(out, "\n"+c.ind)
				pos = stringWidth(c.ind)
			}
		}
	}

	// Flush any line suffixes left at end of document.
	if len(lineSuffixes) > 0 {
		for _, ls := range lineSuffixes {
			out = append(out, PrintDocToString(ls.doc, width))
		}
	}

	return strings.Join(out, "")
}

// chooseExpanded implements conditionalGroup: try each expanded state flat,
// least-expanded first, taking the first that fits; fall back to the most
// expanded (broken).
//
// No construct builds an ExpandedStates group yet, so this is currently
// unreachable. The loop starts at index 1 because, per Prettier, ExpandedStates[0]
// is the group's own Contents - already measured by the caller's flat-fit check
// before chooseExpanded is reached. Revisit this pairing when the first
// conditionalGroup constructor is added.
func chooseExpanded(g *Group, ind string, rest []cmd, rem int, hasLS bool, gmm map[GroupID]mode) cmd {
	if g.Break {
		return cmd{ind, modeBreak, g.ExpandedStates[len(g.ExpandedStates)-1]}
	}
	for i := 1; i < len(g.ExpandedStates); i++ {
		state := g.ExpandedStates[i]
		next := cmd{ind, modeFlat, state}
		if fits(next, rest, rem, hasLS, gmm, false) {
			return next
		}
	}
	return cmd{ind, modeBreak, g.ExpandedStates[len(g.ExpandedStates)-1]}
}

// fits reports whether next, followed by the rest of the current line, fits in
// width columns. It scans in flat mode, stops at the first hard/break line
// (which ends the line, so it fits), and returns false as soon as width goes
// negative. Crucially it measures next AND the remaining stack, not just next.
func fits(next cmd, rest []cmd, width int, hasLineSuffix bool, gmm map[GroupID]mode, mustBeFlat bool) bool {
	restIdx := len(rest)
	cmds := []cmd{next}
	var out []string

	for width >= 0 {
		if len(cmds) == 0 {
			if restIdx == 0 {
				return true
			}
			restIdx--
			cmds = append(cmds, rest[restIdx])
			continue
		}

		c := cmds[len(cmds)-1]
		cmds = cmds[:len(cmds)-1]

		switch d := c.doc.(type) {
		case Text:
			out = append(out, d.S)
			width -= stringWidth(d.S)

		case Concat:
			for i := len(d.Parts) - 1; i >= 0; i-- {
				cmds = append(cmds, cmd{c.ind, c.mode, d.Parts[i]})
			}

		case Fill:
			for i := len(d.Parts) - 1; i >= 0; i-- {
				cmds = append(cmds, cmd{c.ind, c.mode, d.Parts[i]})
			}

		case Indent:
			cmds = append(cmds, cmd{c.ind, c.mode, d.Contents})

		case Align:
			cmds = append(cmds, cmd{c.ind, c.mode, d.Contents})

		case IndentIfBreak:
			cmds = append(cmds, cmd{c.ind, c.mode, d.Contents})

		case Trim:
			width += trimTrailing(&out)

		case *Group:
			if mustBeFlat && d.Break {
				return false
			}
			gm := c.mode
			if d.Break {
				gm = modeBreak
			}
			contents := d.Contents
			if d.ExpandedStates != nil && gm == modeBreak {
				contents = d.ExpandedStates[len(d.ExpandedStates)-1]
			}
			cmds = append(cmds, cmd{c.ind, gm, contents})

		case IfBreak:
			gm := c.mode
			if d.GroupID != 0 {
				gm = gmm[d.GroupID]
			}
			chosen := d.FlatContents
			if gm == modeBreak {
				chosen = d.BreakContents
			}
			if chosen != nil {
				cmds = append(cmds, cmd{c.ind, c.mode, chosen})
			}

		case Line:
			if c.mode == modeBreak || d.Hard {
				return true
			}
			if !d.Soft {
				out = append(out, " ")
				width--
			}

		case LineSuffix:
			hasLineSuffix = true

		case LineSuffixBoundary:
			if hasLineSuffix {
				return false
			}

		case BreakParent:
			// no-op
		}
	}

	return false
}

// trimTrailing strips trailing spaces/tabs from the already-emitted output's
// last segment(s) and returns how many characters were removed (used to correct
// the column counter).
func trimTrailing(out *[]string) int {
	o := *out
	trimmed := 0
	for len(o) > 0 {
		last := o[len(o)-1]
		stripped := strings.TrimRight(last, " \t")
		if stripped == last {
			break
		}
		trimmed += len(last) - len(stripped)
		if stripped == "" {
			o = o[:len(o)-1]
			continue
		}
		o[len(o)-1] = stripped
		break
	}
	*out = o
	return trimmed
}

// propagateBreaks marks every Group that (transitively) contains a hard break or
// BreakParent as Break=true, so those groups render broken without measuring. It
// runs once before PrintDocToString and mutates *Group nodes in place.
func propagateBreaks(d Doc) (forced bool) {
	switch n := d.(type) {
	case BreakParent:
		return true
	case Line:
		return n.Hard
	case *Group:
		child := propagateBreaks(n.Contents)
		for _, st := range n.ExpandedStates {
			if propagateBreaks(st) {
				child = true
			}
		}
		if child {
			n.Break = true
		}
		// A broken group still propagates a forced break to its parents only
		// when it actually contains a hard break - matching Prettier, an
		// explicit shouldBreak does not force ancestors. So return child here.
		return child
	case Concat:
		for _, p := range n.Parts {
			if propagateBreaks(p) {
				forced = true
			}
		}
		return forced
	case Fill:
		for _, p := range n.Parts {
			if propagateBreaks(p) {
				forced = true
			}
		}
		return forced
	case Indent:
		return propagateBreaks(n.Contents)
	case Align:
		return propagateBreaks(n.Contents)
	case IndentIfBreak:
		return propagateBreaks(n.Contents)
	case IfBreak:
		a := propagateBreaks(n.BreakContents)
		b := propagateBreaks(n.FlatContents)
		return a || b
	case LineSuffix:
		// A break inside a line suffix should not force the enclosing group:
		// suffixes are deferred and don't affect the current line's fit.
		propagateBreaks(n.Contents)
		return false
	}
	return false
}
